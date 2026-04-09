package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/Jingqi0327/GeeCache/geecache"
	"github.com/Jingqi0327/GeeCache/geecache/registry"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

var defaultTTL = 100 * time.Second
var gcInterval = 300 * time.Second

var registryClient registry.Registry

func createGroup() *geecache.Group {
	return geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("[SlowDB] searching key: %s", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}), defaultTTL, gcInterval)
}

func startCacheServer(registry registry.Registry, addr string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)
	gee.RegisterPeers(peers)

	// 启动节点发现，先获取etcd中已有的节点，并更新到peers中
	go registry.Discovery(func(addrs ...string) {
		peers.Set(addrs...)
	})

	// 注册节点
	stopCh := make(chan error)
	go func() {
		err := registry.Register(addr, stopCh)
		if err != nil {
			log.Fatal(err)
		}
	}()

	// 启动节点
	log.Println("geecache is running at", addr)
	err := http.ListenAndServe(addr[7:], peers)
	if err != nil {
		stopCh <- err
	}
}

func startAPIServer(apiAddr string, gee *geecache.Group) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		view, err := gee.Get(key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write(view.ByteSlice())
	}

	http.Handle("/api", http.HandlerFunc(handler))

	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func setRegistry(name string) {
	switch name {
	case "etcd":
		var err error
		registryClient, err = registry.NewEtcdRegistry([]string{"127.0.0.1:2379"}, "/geecache/")
		if err != nil {
			log.Fatal(err)
		}
	case "gee":
		registryClient = registry.NewGeeRegistry("http://localhost:8000/_gee_/registry", 3*time.Second)
	}
}

func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":8000")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func main() {
	var port int
	var api bool
	var geeRegistry bool
	var geeRegistryServer bool
	// 这边注册了两个命令行参数
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.BoolVar(&geeRegistry, "geeRegistry", true, "Start a Geeregistry ?")
	flag.BoolVar(&geeRegistryServer, "geeRegistryServer", false, "Start a GeeregistryServer ?")
	flag.Parse()

	apiAddr := "http://localhost:9999"

	var wg sync.WaitGroup
	if geeRegistryServer {
		wg.Add(1)
		go startRegistry(&wg)
	}
	wg.Wait()

	if geeRegistry {
		setRegistry("gee")
	} else {
		setRegistry("etcd")
	}

	addr := fmt.Sprintf("http://localhost:%d", port)

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(registryClient, addr, gee)
}
