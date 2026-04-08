package geecache

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGet(t *testing.T) {
	// 统计从数据库中获取数据的次数
	loadCounts := make(map[string]int, len(db))
	gee := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("[SlowDB] search key %s\n", key)
			if v, ok := db[key]; ok { // 从db中获取数据
				// 统计从数据库中获取数据的次数
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key] += 1
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), 1*time.Minute, 1*time.Minute)
	// 遍历db中的数据
	for k, v := range db {
		// 第一次访问，应该从数据库中获取数据
		byteView, err := gee.Get(k)
		require.NoError(t, err)
		require.Equal(t, ByteView{b: []byte(v)}, byteView)
		// 第二次访问，应该命中缓存
		byteView, err = gee.Get(k)
		require.NoError(t, err)
		require.Equal(t, ByteView{b: []byte(v)}, byteView)
		require.Equal(t, 1, loadCounts[k], "cache miss count should be 1")
	}
}
