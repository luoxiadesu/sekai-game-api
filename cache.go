package main

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value     []byte
	expiresAt time.Time
}

type TTLCache struct {
	ttl  time.Duration
	mu   sync.RWMutex
	data map[string]cacheEntry
}

func NewTTLCache(ttl time.Duration) *TTLCache {
	return &TTLCache{ttl: ttl, data: make(map[string]cacheEntry)}
}

func (c *TTLCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, ok := c.data[key]
	c.mu.RUnlock()
	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.value, true
}

func (c *TTLCache) Set(key string, value []byte) {
	c.mu.Lock()
	c.data[key] = cacheEntry{value: value, expiresAt: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}
