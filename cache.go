package geecache

import (
	"sync"

	"github.com/Jingqi0327/GeeCache/lru"
)

type cache struct {
	mu         sync.Mutex // 保证不同Goroutine不会同时操作Cache
	lru        *lru.Cache
	cacheBytes int64 // 允许使用的最大内存
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lru == nil {
		return
	}
	
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
