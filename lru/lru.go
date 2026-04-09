package lru

import (
	"container/list"
	"log"
	"time"
)

// 定义一个LRUCache
type Cache struct {
	maxBytes  int64
	nBytes    int64
	ll        *list.List
	cache     map[string]*list.Element
	OnEvicted func(key string, value Value)
}

// Cache中存储的一种元素
type entry struct {
	key      string
	value    Value
	deadline time.Time
}

// 判断是否过期
func (e *entry) isExpired() bool {
	if e.deadline.IsZero() {
		return false
	}
	return time.Now().After(e.deadline)
}

// Value 接口,想要存到Cache里的数据必须实现这个接口
type Value interface {
	Len() int
}

// 实例化一个LRUCache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 获取Cache中的一个数据
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		kv := ele.Value.(*entry)

		// 判断是否过期
		if kv.isExpired() {
			log.Printf("[GeeCache] Remove expired key: %s\n", kv.key)
			c.RemoveElement(ele)
			return nil, false
		}

		c.ll.MoveToFront(ele)
		return kv.value, true
	}
	return
}

// 向缓存中添加/更新一个数据
func (c *Cache) Add(key string, value Value, duration time.Duration) {
	var deadline time.Time
	if duration != 0 {
		deadline = time.Now().Add(duration)
	}

	if ele, ok := c.cache[key]; ok {
		// 修改Cache中某个key的value
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
		kv.deadline = deadline
	} else {
		// 在Cache中插入一个新的数据
		ele := c.ll.PushFront(&entry{key, value, deadline})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	// 最后检查占用内存有没有超,超了就删除最近未使用的
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}

// 从缓存中删除某个元素
func (c *Cache) RemoveElement(ele *list.Element) {
	if ele != nil {
		// 从链表中删除该节点
		c.ll.Remove(ele)
		// 获取链表中节点的值并转成entry类型
		entry := ele.Value.(*entry)
		// 从cache的map中也删除这个键值对
		delete(c.cache, entry.key)
		// 更新内存占用
		c.nBytes -= int64(len(entry.key)) + int64(entry.value.Len())
		// 执行回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(entry.key, entry.value)
		}
	}
}

// 删除最近最少访问的数据
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	c.RemoveElement(ele)
}

// 删除所有过期的元素
func (c *Cache) RemoveExpired() {
	if c.ll == nil {
		return
	}
	var prev *list.Element
	for e := c.ll.Back(); e != nil; e = prev {
		prev = e.Prev()
		kv := e.Value.(*entry)
		if kv.isExpired() {
			log.Printf("[GeeCache] Remove expired key: %s\n", kv.key)
			c.RemoveElement(e)
		}
	}
}

// 返回缓存中元素的个数
func (c *Cache) Len() int {
	return c.ll.Len()
}
