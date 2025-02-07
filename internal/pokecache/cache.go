package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	items map[string]cacheEntry
	mu    sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]cacheEntry),
	}
	cache.reapLoop(interval)
	return cache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	c.items[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.mu.Unlock()
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	val, ok := c.items[key]
	c.mu.Unlock()
	if !ok {
		return nil, false
	}
	return val.val, true
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			<-ticker.C
			c.mu.Lock()
			for key, entry := range c.items {
				if time.Since(entry.createdAt) > interval {
					delete(c.items, key)
				}
			}
			c.mu.Unlock()
		}
	}()
}
