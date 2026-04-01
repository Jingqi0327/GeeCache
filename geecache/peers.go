package geecache

type PeerPicker interface {
	// PickPeer()方法根据传入的key选择相应的节点 PeerGetter
	PickPeer(key string) (PeerGetter, bool)
}

type PeerGetter interface {
	// Get()方法从对应group查找缓存值
	Get(group string, key string) ([]byte, error)
}
