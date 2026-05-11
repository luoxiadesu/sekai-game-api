package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// handleRanking: GET /api/{region}/event/{event_id}/ranking
// 多上游并发抢答，每个上游内部并行取 ranking-top100 + ranking-border，合并成
// bot 期望的 {top100, border} 结构。
func (s *Server) handleRanking(w http.ResponseWriter, r *http.Request, region, eventIDStr string) {
	eventID, err := strconv.Atoi(eventIDStr)
	if err != nil || eventID <= 0 {
		writeDetailMsg(w, http.StatusBadRequest, "invalid event_id")
		return
	}

	cacheKey := "ranking:" + region + ":" + eventIDStr
	if cached, ok := s.cache.Get(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "hit")
		_, _ = w.Write(cached)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(s.cfg.HTTP.TimeoutSeconds+2)*time.Second)
	defer cancel()

	body, source, err := s.fetchRankingHedged(ctx, region, eventID)
	if err != nil {
		writeDetailMsg(w, http.StatusBadGateway, fmt.Sprintf("upstream failed: %v", err))
		return
	}

	s.cache.Set(cacheKey, body)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "miss")
	w.Header().Set("X-Source", source)
	_, _ = w.Write(body)
}

// fetchRankingHedged: 每个上游需要并行发两条请求 (top100 + border) 再合并 —
// 不是简单代理，所以不复用通用 Hedged 框架。
func (s *Server) fetchRankingHedged(ctx context.Context, region string, eventID int) ([]byte, string, error) {
	n := len(s.upstreams)
	if n == 0 {
		return nil, "", fmt.Errorf("no upstreams configured")
	}

	raceCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type item struct {
		body   []byte
		source string
		err    error
	}
	out := make(chan item, n)

	for _, up := range s.upstreams {
		go func(u *Upstream) {
			merged, err := fetchRankingOneUpstream(raceCtx, u, region, eventID)
			if err != nil {
				out <- item{err: fmt.Errorf("%s: %w", u.Name(), err)}
				return
			}
			raw, err := json.Marshal(merged)
			if err != nil {
				out <- item{err: fmt.Errorf("%s marshal: %w", u.Name(), err)}
				return
			}
			out <- item{body: raw, source: u.Name()}
		}(up)
	}

	var errs []string
	for i := 0; i < n; i++ {
		it := <-out
		if it.err == nil {
			return it.body, it.source, nil
		}
		errs = append(errs, it.err.Error())
	}
	return nil, "", fmt.Errorf("all upstreams failed: %v", errs)
}

func fetchRankingOneUpstream(ctx context.Context, u *Upstream, region string, eventID int) (map[string]any, error) {
	var top100, border map[string]any
	var topErr, borderErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		top100, topErr = fetchJSONFromUpstream(ctx, u, fmt.Sprintf("%s/%s/event/%d/ranking-top100", u.cfg.Base, region, eventID))
	}()
	go func() {
		defer wg.Done()
		border, borderErr = fetchJSONFromUpstream(ctx, u, fmt.Sprintf("%s/%s/event/%d/ranking-border", u.cfg.Base, region, eventID))
	}()
	wg.Wait()
	if topErr != nil {
		return nil, fmt.Errorf("top100: %w", topErr)
	}
	if borderErr != nil {
		return nil, fmt.Errorf("border: %w", borderErr)
	}
	return map[string]any{"top100": top100, "border": border}, nil
}

func fetchJSONFromUpstream(ctx context.Context, u *Upstream, url string) (map[string]any, error) {
	data, _, err := u.Do(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return out, nil
}
