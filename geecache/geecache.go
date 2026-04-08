package geecache

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/Jingqi0327/GeeCache/geecache/singleflight"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name       string // Group的名字
	getter     Getter // 缓存数据未命中时回调函数
	mainCache  cache  // 支持并发的Cache
	peers      PeerPicker
	loader     *singleflight.Group
	defaultTTL time.Duration // 默认缓存时间
	gcInterval time.Duration // 过期缓存清理时间
}

// 存放所有Group的全局变量
var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter, defaultTTL time.Duration, gcInterval time.Duration) *Group {
	if getter == nil {
		panic("nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:       name,
		getter:     getter,
		mainCache:  cache{cacheBytes: cacheBytes},
		loader:     &singleflight.Group{},
		defaultTTL: defaultTTL,
		gcInterval: gcInterval,
	}
	g.mainCache.startGC(gcInterval)
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	defer mu.RUnlock()
	g := groups[name]
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("[GeeCache] hit : %s\n", key)
		return v, nil
	}

	return g.load(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	fn := func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err := g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	}

	viewi, err := g.loader.Do(key, fn)
	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value, withJitter(g.defaultTTL))
}

// 添加随机抖动，防止缓存同时过期
func withJitter(duration time.Duration) time.Duration {
	jitterRange := int64(duration) / 10
	if jitterRange == 0 {
		return duration
	}
	jitter := rand.Int63n(jitterRange)
	return duration + time.Duration(jitter)
}
