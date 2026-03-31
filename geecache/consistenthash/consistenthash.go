package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // Hash函数
	replicas int            // 虚拟节点的倍数
	keys     []int          // 排序过的hash环,通过排序来实现"环"的效果
	hashMap  map[int]string //虚拟节点与真实节点的映射
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// 这边传入的是节点名
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			// 以i+key作为虚拟节点的名称
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// 这边传入的是缓存数据的key
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	// 计算key的hash值
	hash := int(m.hash([]byte(key)))
	// 在hash环上找第一个比key的hash值大的虚拟节点
	// 相当于是顺时针找到的第一个节点
	idx:= sort.Search(len(m.keys),func(i int) bool {
		return m.keys[i]>=hash
	})

	// 假如key的hash比里面的虚拟节点的hash都大
	// idx就是len(m.keys),那么顺时针最靠近的节点就是m.keys[0]
	return m.hashMap[m.keys[idx%len(m.keys)]]
}
