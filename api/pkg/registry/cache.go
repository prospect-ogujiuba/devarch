package registry

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
}

func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]*cacheEntry),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	entry, ok := c.entries[key]
	c.mu.RUnlock()

	if !ok || time.Now().After(entry.expiresAt) {
		return nil, false
	}
	return entry.data, true
}

func (c *Cache) Set(key string, data interface{}, ttl time.Duration) {
	c.mu.Lock()
	c.entries[key] = &cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}
