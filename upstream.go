package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Upstream 封装一个上游服务（haruki / exmeaning）的访问。
type Upstream struct {
	cfg    UpstreamConfig
	client *http.Client
}

func NewUpstream(cfg UpstreamConfig, timeout time.Duration) *Upstream {
	return &Upstream{cfg: cfg, client: &http.Client{Timeout: timeout}}
}

func (u *Upstream) Name() string { return u.cfg.Name }

// Do 发起请求并返回原始 (body, contentType)。token 为空则不附鉴权头。
func (u *Upstream) Do(ctx context.Context, method, url string, body io.Reader) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, "", err
	}
	if token := u.cfg.Token(); token != "" && u.cfg.AuthHeader != "" {
		req.Header.Set(u.cfg.AuthHeader, token)
	}
	req.Header.Set("User-Agent", "sekai-game-api/1.0")

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode != http.StatusOK {
		snippet := string(data)
		if len(snippet) > 256 {
			snippet = snippet[:256]
		}
		return nil, "", fmt.Errorf("upstream %s status %d: %s", u.cfg.Name, resp.StatusCode, snippet)
	}
	return data, resp.Header.Get("Content-Type"), nil
}

// HedgedResult 是 Hedged 抢答的返回结构。
type HedgedResult struct {
	Body        []byte
	ContentType string
	Source      string
}

// Hedged 并发向所有上游发请求，首个成功立即返回并取消其他。
// build 用于为每个上游生成请求 (method, url, body) 三元组。
func Hedged(
	ctx context.Context,
	upstreams []*Upstream,
	build func(*Upstream) (method, url string, body io.Reader, err error),
) (*HedgedResult, error) {
	n := len(upstreams)
	if n == 0 {
		return nil, fmt.Errorf("no upstreams configured")
	}

	raceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type item struct {
		res *HedgedResult
		err error
	}
	out := make(chan item, n)

	var wg sync.WaitGroup
	wg.Add(n)
	for _, up := range upstreams {
		go func(u *Upstream) {
			defer wg.Done()
			method, url, body, err := build(u)
			if err != nil {
				out <- item{err: fmt.Errorf("%s build: %w", u.Name(), err)}
				return
			}
			data, ct, err := u.Do(raceCtx, method, url, body)
			if err != nil {
				out <- item{err: fmt.Errorf("%s: %w", u.Name(), err)}
				return
			}
			out <- item{res: &HedgedResult{Body: data, ContentType: ct, Source: u.Name()}}
		}(up)
	}

	var errs []string
	for i := 0; i < n; i++ {
		it := <-out
		if it.err == nil {
			return it.res, nil
		}
		log.Printf("upstream failed: %v", it.err)
		errs = append(errs, it.err.Error())
	}
	return nil, fmt.Errorf("all upstreams failed: %s", strings.Join(errs, "; "))
}
