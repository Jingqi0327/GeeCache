package registry

type Registry interface {
	Register(addr string, stop chan error) error
	Discovery(updatePeers func(peers ...string))
	Close()
}