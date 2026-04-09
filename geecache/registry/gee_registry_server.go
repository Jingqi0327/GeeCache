package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type GeeRegistryServer struct {
	timeout time.Duration
	mu      sync.Mutex // protect following
	servers map[string]*Peer
}

type Peer struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_gee_/registry"
	defaultTimeout = time.Second * 5
)

func New(timeout time.Duration) *GeeRegistryServer {
	return &GeeRegistryServer{
		servers: make(map[string]*Peer),
		timeout: timeout,
	}
}

var DefaultGeeRegister = New(defaultTimeout)

func (r *GeeRegistryServer) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	s := r.servers[addr]
	if s == nil {
		r.servers[addr] = &Peer{Addr: addr, start: time.Now()}
	} else {
		s.start = time.Now() // if exists, update start time to keep alive
	}
}

func (r *GeeRegistryServer) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	var alive []string
	for addr, s := range r.servers {
		// 是否有设超时时间 或 server的时间加上超时时间有没有超过现在的时间
		if r.timeout == 0 || s.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

// Runs at /_gee_/registry
func (r *GeeRegistryServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case "GET":
		// keep it simple, server is in req.Header
		w.Header().Set("X-Geerpc-Servers", strings.Join(r.aliveServers(), ","))
	case "POST":
		// keep it simple, server is in req.Header
		addr := req.Header.Get("X-Geerpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// HandleHTTP registers an HTTP handler for GeeRegistryServer messages on registryPath
func (r *GeeRegistryServer) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("gee registry path:", registryPath)
}

func HandleHTTP() {
	DefaultGeeRegister.HandleHTTP(defaultPath)
}
