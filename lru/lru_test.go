package lru

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)


type String string

func (s String) Len() int {
	return len(s)
}

func TestAdd(t *testing.T) {
	lru := New(int64(0), nil)
	// 添加数据
	lru.Add("key1", String("value1"), 0)
	lru.Add("key2", String("value2"), 0)
	require.Equal(t, 2, lru.Len())
	require.Equal(t, lru.nBytes, int64(len("key1")+len("value1")+len("key2")+len("value2")))
	// 更新数据
	lru.Add("key1", String("value111"), 0)
	require.Equal(t, 2, lru.Len())
	require.Equal(t, lru.nBytes, int64(len("key1")+len("value111")+len("key2")+len("value2")))
}

func TestGet(t *testing.T) {
	lru := New(int64(0), nil)
	lru.Add("key1", String("value1"), 0)
	value, ok := lru.Get("key1")
	require.True(t, ok)
	require.Equal(t, value, String("value1"))
	_, ok = lru.Get("key2")
	require.False(t, ok)
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3, k4 := "key1", "key2", "key3", "key4"
	v1, v2, v3, v4 := String("value1"), String("value2"), String("value3"), String("value4")
	cap := len(k1) + len(v1) + len(k2) + len(v2)
	lru := New(int64(cap), nil)
	lru.Add(k1, v1, 0)
	lru.Add(k2, v2, 0)
	lru.Add(k3, v3, 0)
	_, ok := lru.Get(k1)
	require.False(t, ok)
	_, ok = lru.Get(k2)
	require.True(t, ok)
	_, ok = lru.Get(k3)
	require.True(t, ok)
	lru.Add(k4, v4, 0)
	_, ok = lru.Get(k2)
	require.False(t, ok)

}

func TestGetExpired(t *testing.T) {
	k1 := "key1"
	v1 := String("value1")
	lru := New(int64(0), nil)
	lru.Add(k1, v1, -1*time.Second)
	_, ok := lru.Get(k1)
	require.False(t, ok)
}

func TestRemoveExpired(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := String("value1"), String("value2"), String("value3")
	lru := New(int64(0), nil)
	lru.Add(k1, v1, -1*time.Second)
	lru.Add(k2, v2, -1*time.Second)
	lru.Add(k3, v3, 0)
	lru.RemoveExpired()
	_, ok := lru.Get(k1)
	require.False(t, ok)
	_, ok = lru.Get(k2)
	require.False(t, ok)
	_, ok = lru.Get(k3)
	require.True(t, ok)
}


func TestOnEvicted(t *testing.T) {
	keys := []string{}
	onEvicted := func(key string, value Value) {
		keys = append(keys, key)
	}
	k1, k2, k3 := "key1", "key2", "key3"
	v1, v2, v3 := String("value1"), String("value2"), String("value3")
	cap := len(k1) + len(v1)
	lru := New(int64(cap), onEvicted)
	lru.Add(k1, v1, 0)
	lru.Add(k2, v2, 0)
	lru.Add(k3, v3, 0)
	require.Equal(t, []string{k1, k2}, keys)
}