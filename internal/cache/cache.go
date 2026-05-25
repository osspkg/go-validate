package cache

import "sync"

type Cache[K comparable, T any] struct {
	data map[K]T
	mux  sync.RWMutex
}

func New[K comparable, T any]() *Cache[K, T] {
	return &Cache[K, T]{
		data: make(map[K]T, 10),
	}
}

func (c *Cache[K, T]) Get(key K) (T, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	val, ok := c.data[key]
	return val, ok
}

func (c *Cache[K, T]) Set(key K, val T) {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.data[key] = val
}
