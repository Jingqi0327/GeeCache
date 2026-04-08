package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Jingqi0327/GeeCache/geecache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

var defaultTTL = 10 * time.Second
var gcInterval = 30 * time.Second

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

func startCacheServer(addr string, addrs []string, gee *geecache.Group) {
	peers := geecache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gee.RegisterPeers(peers)

	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
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

func main() {
	var port int
	var api bool
	// 这边注册了两个命令行参数
	flag.IntVar(&port, "port", 8001, "Geecache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, addr := range addrMap {
		addrs = append(addrs, addr)
	}

	gee := createGroup()
	if api {
		go startAPIServer(apiAddr, gee)
	}
	startCacheServer(addrMap[port], addrs, gee)
}
