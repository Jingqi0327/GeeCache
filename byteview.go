package geecache

// 创建一个统一的数据类型，存储到Cache中
type ByteView struct {
	b []byte
}

// ByteView实现了Value接口
func (v ByteView) Len() int {
	return len(v.b)
}

// 返回一个只读的Clone，防止外部修改ByteView中的数据
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 返回一个字符串形式的ByteView
func (v ByteView) String() string {
	return string(v.b)
}

func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}