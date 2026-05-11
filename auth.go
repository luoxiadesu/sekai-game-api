package main

import (
	"net/http"
	"strings"
)

// checkAuth 与 bot gameapi.py 的鉴权约定一致: Authorization: Bearer {token}
func checkAuth(cfg *Config, r *http.Request) bool {
	if !cfg.Server.AuthEnabled {
		return true
	}
	want := strings.TrimSpace(cfg.Server.BearerToken)
	if want == "" {
		return false
	}
	got := strings.TrimSpace(r.Header.Get("Authorization"))
	return got == "Bearer "+want
}
