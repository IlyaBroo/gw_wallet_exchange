package cache

import (
	"sync"
	"time"
)

type Cache struct {
	mu           sync.RWMutex
	data         map[string]float32
	specialRates map[string]float32
	ttl          time.Duration
}

func NewCache(ttl time.Duration) *Cache {
	cache := new(Cache)
	cache.data = make(map[string]float32)
	cache.specialRates = make(map[string]float32)
	cache.ttl = ttl
	return cache
}

func (c *Cache) GetAll() map[string]float32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	copyData := make(map[string]float32)
	for k, v := range c.data {
		copyData[k] = v
	}
	return copyData
}

func (c *Cache) Set(data map[string]float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, value := range data {
		c.data[key] = value
	}

	time.AfterFunc(c.ttl, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		c.data = make(map[string]float32)
	})
}

func (c *Cache) GetSpecificRate(key string) (float32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, found := c.specialRates[key]
	return val, found
}

func (c *Cache) SetSpecificRate(key string, value float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.specialRates[key] = value

	time.AfterFunc(c.ttl, func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.specialRates, key)
	})
}
