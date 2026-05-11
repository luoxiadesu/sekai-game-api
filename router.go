package main

import (
	"net/http"
	"strings"
	"time"
)

type Server struct {
	cfg       *Config
	upstreams []*Upstream
	cache     *TTLCache
}

func NewServer(cfg *Config) *Server {
	timeout := time.Duration(cfg.HTTP.TimeoutSeconds) * time.Second
	ups := make([]*Upstream, 0, len(cfg.Upstreams))
	for _, u := range cfg.Upstreams {
		ups = append(ups, NewUpstream(u, timeout))
	}
	return &Server{
		cfg:       cfg,
		upstreams: ups,
		cache:     NewTTLCache(time.Duration(cfg.Cache.TTLSeconds) * time.Second),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("/api/", s.dispatch)
	return mux
}

// dispatch 解析 /api/... 路径并分发到具体 handler。
// bot 的 Bearer 鉴权统一在此校验。
func (s *Server) dispatch(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(s.cfg, r) {
		writeDetailMsg(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// parts[0] == "api"

	// /api/status
	if len(parts) == 2 && parts[1] == "status" && r.Method == http.MethodGet {
		s.handleStatus(w, r)
		return
	}

	if len(parts) < 3 {
		writeDetailMsg(w, http.StatusNotFound, "not found")
		return
	}
	region := parts[1]

	// /api/{region}/event/{event_id}/ranking
	if len(parts) == 5 && parts[2] == "event" && parts[4] == "ranking" && r.Method == http.MethodGet {
		s.handleRanking(w, r, region, parts[3])
		return
	}

	// /api/{region}/user/{uid}/{kind}
	if len(parts) == 5 && parts[2] == "user" {
		uid, kind := parts[3], parts[4]
		switch {
		case kind == "profile" && r.Method == http.MethodGet:
			s.handleProfile(w, r, region, uid)
			return
		case kind == "suite" && r.Method == http.MethodGet:
			s.handleStubSuite(w, r, region, uid)
			return
		case kind == "mysekai" && r.Method == http.MethodGet:
			s.handleStubMysekai(w, r, region, uid)
			return
		case kind == "send_boost" && r.Method == http.MethodPost:
			s.handleStubSendBoost(w, r, region, uid)
			return
		case kind == "ad_result" && r.Method == http.MethodGet:
			s.handleStubAdResult(w, r, region, uid)
			return
		}
	}

	// /api/{region}/mysekai/{action}
	if len(parts) == 4 && parts[2] == "mysekai" {
		switch parts[3] {
		case "photo":
			if r.Method == http.MethodPost {
				s.handlePhoto(w, r, region)
				return
			}
		case "upload_time":
			if r.Method == http.MethodPost {
				s.handleStubUploadTime(w, r, region)
				return
			}
		case "subscriptions":
			if r.Method == http.MethodPut {
				s.handleStubSubscriptions(w, r, region)
				return
			}
		}
	}

	// /api/{region}/create_account
	if len(parts) == 3 && parts[2] == "create_account" && r.Method == http.MethodPost {
		s.handleStubCreateAccount(w, r, region)
		return
	}

	// /api/{region}/ad_result/update_time
	if len(parts) == 4 && parts[2] == "ad_result" && parts[3] == "update_time" && r.Method == http.MethodGet {
		s.handleStubAdResultUpdateTime(w, r, region)
		return
	}

	writeDetailMsg(w, http.StatusNotFound, "not found")
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
