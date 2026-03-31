package consistenthash

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHashing(t *testing.T) {
	// 为了测试方便,自定义个伪hash函数,将传入的key按十进制解析并转成uint32型
	hashFunc := func(key []byte) uint32 {
		i, _ := strconv.Atoi(string(key))
		return uint32(i)
	}

	hash := New(3, hashFunc)

	// "2": 2,12,22
	// "4": 4,14,24
	// "6": 6,16,26
	hash.Add("2", "4", "6")

	testCases := map[string]string{
		// 请求的key 与 实际服务的节点
		"2":  "2",
		"11": "2",
		"23": "4",
		"27": "2",
	}

	for k, v := range testCases {
		// Get()出的节点要与预期节点一样
		require.Equal(t, hash.Get(k), v)
	}

	// 添加一个节点
	// "8": 8,18,28
	hash.Add("8")
	// 此时key=27 应由节点8服务
	testCases["27"]="8"

	for k, v := range testCases {
		// Get()出的节点要与预期节点一样
		require.Equal(t, hash.Get(k), v)
	}
}
