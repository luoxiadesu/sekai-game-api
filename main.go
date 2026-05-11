package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func main() {
	configPath := flag.String("config", "config.yaml", "config file path")
	flag.Parse()

	cfg, err := LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	for _, u := range cfg.Upstreams {
		if u.TokenEnv != "" && u.Token() == "" {
			log.Printf("warning: upstream %s token env %s not set", u.Name, u.TokenEnv)
		}
	}
	if !cfg.Server.AuthEnabled {
		log.Printf("warning: auth_enabled=false — gateway is unauthenticated")
	} else if cfg.Server.BearerToken == "" {
		log.Printf("warning: auth_enabled=true but bearer_token empty — all requests will be rejected")
	}

	srv := NewServer(cfg)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("sekai-game-api listening on %s, %d upstreams, ttl=%ds",
		addr, len(cfg.Upstreams), cfg.Cache.TTLSeconds)
	if err := http.ListenAndServe(addr, srv.Handler()); err != nil {
		log.Fatal(err)
	}
}
