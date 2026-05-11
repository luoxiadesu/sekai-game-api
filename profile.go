package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// handleProfile: GET /api/{region}/user/{uid}/profile[?use_cache=...]
// 透传到上游 {base}/{region}/{uid}/profile，hedged 抢答。
func (s *Server) handleProfile(w http.ResponseWriter, r *http.Request, region, uid string) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(s.cfg.HTTP.TimeoutSeconds+2)*time.Second)
	defer cancel()

	query := r.URL.RawQuery

	res, err := Hedged(ctx, s.upstreams, func(u *Upstream) (string, string, io.Reader, error) {
		url := fmt.Sprintf("%s/%s/%s/profile", u.cfg.Base, region, uid)
		if query != "" {
			url += "?" + query
		}
		return http.MethodGet, url, nil, nil
	})
	if err != nil {
		writeDetailMsg(w, http.StatusBadGateway, fmt.Sprintf("upstream failed: %v", err))
		return
	}

	ct := res.ContentType
	if ct == "" {
		ct = "application/json"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Source", res.Source)
	_, _ = w.Write(res.Body)
}
