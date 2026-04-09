package geecache

import (
	"log"
	"sync"
	"time"

	"github.com/Jingqi0327/GeeCache/lru"
)

type cache struct {
	mu         sync.Mutex // 保证不同Goroutine不会同时操作Cache
	lru        *lru.Cache
	cacheBytes int64 // 允许使用的最大内存
}

func (c *cache) add(key string, value ByteView, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value, duration)
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

func (c *cache) startGC(gcInterval time.Duration) {
	ticker := time.NewTicker(gcInterval)
	go func() {
		for {
			<-ticker.C
			c.mu.Lock()
			log.Printf("[GeeCache] GC start...\n")
			if c.lru != nil {
				c.lru.RemoveExpired()
			}
			log.Printf("[GeeCache] GC done...\n")
			c.mu.Unlock()
		}
	}()
}
