package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port        int    `yaml:"port"`
		AuthEnabled bool   `yaml:"auth_enabled"`
		BearerToken string `yaml:"bearer_token"`
	} `yaml:"server"`

	Cache struct {
		TTLSeconds int `yaml:"ttl_seconds"`
	} `yaml:"cache"`

	HTTP struct {
		TimeoutSeconds int `yaml:"timeout_seconds"`
	} `yaml:"http"`

	Upstreams []UpstreamConfig `yaml:"upstreams"`
}

type UpstreamConfig struct {
	Name       string `yaml:"name"`
	Base       string `yaml:"base"`
	ImageBase  string `yaml:"image_base"`
	AuthHeader string `yaml:"auth_header"`
	TokenEnv   string `yaml:"token_env"`
}

func (u UpstreamConfig) Token() string { return os.Getenv(u.TokenEnv) }

func LoadConfig(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	cfg := &Config{}
	if err := yaml.Unmarshal(raw, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 11000
	}
	if cfg.Cache.TTLSeconds == 0 {
		cfg.Cache.TTLSeconds = 5
	}
	if cfg.HTTP.TimeoutSeconds == 0 {
		cfg.HTTP.TimeoutSeconds = 10
	}
	if v := strings.TrimSpace(os.Getenv("GAMEAPI_BEARER_TOKEN")); v != "" {
		cfg.Server.BearerToken = v
	}
	for i := range cfg.Upstreams {
		cfg.Upstreams[i].Base = strings.TrimRight(cfg.Upstreams[i].Base, "/")
		cfg.Upstreams[i].ImageBase = strings.TrimRight(cfg.Upstreams[i].ImageBase, "/")
	}
	return cfg, nil
}
