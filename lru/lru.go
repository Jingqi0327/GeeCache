package lru

import "container/list"

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
	key   string
	value Value
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
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

// 删除最近最少访问的数据
func (c *Cache) RemoveOldest() {
	ele:=c.ll.Back()
	if ele!=nil{
		// 从链表中删除该节点
		c.ll.Remove(ele)
		// 获取链表中节点的值并转成entry类型
		entry := ele.Value.(*entry)
		// 从cache的map中也删除这个键值对
		delete(c.cache,entry.key)
		// 更新内存占用
		c.nBytes-=int64(len(entry.key))+int64(entry.value.Len())
		// 执行回调函数
		if c.OnEvicted!=nil{
			c.OnEvicted(entry.key,entry.value)
		}
	}
}

func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 修改Cache中某个key的value
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 在Cache中插入一个新的数据
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	// 最后检查占用内存有没有超,超了就删除最近未使用的
	for c.maxBytes != 0 && c.maxBytes < c.nBytes {
		c.RemoveOldest()
	}
}


func (c *Cache) Len() int{
	return  c.ll.Len()
}