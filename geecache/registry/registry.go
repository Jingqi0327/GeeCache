package registry

type RegistryClient interface {
	Register(addr string, stop chan error) error
	Discovery(updatePeers func(peers ...string))
	Close()
}