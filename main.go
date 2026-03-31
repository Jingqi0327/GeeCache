package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Jingqi0327/GeeCache/geecache"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	geecache.NewGroup("scores", 2<<10, geecache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Printf("[SlowDB] searching key: %s", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := geecache.NewHTTPPool(addr)
	log.Printf("geecache is running at %s", addr)
	log.Fatal(http.ListenAndServe(addr, peers))

}
