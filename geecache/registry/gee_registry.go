package registry

import (
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type GeeRegistry struct {
	registry       string
	updateDuration time.Duration
	peers          []string
}

func NewGeeRegistry(registry string, updateDuration time.Duration) Registry {
	return &GeeRegistry{
		registry:       registry,
		updateDuration: updateDuration,
	}
}

func (d *GeeRegistry) Register(addr string, stop chan error) error {
	// 注册并发送心跳
	Heartbeat(d.registry, addr, 0)
	for err := range stop {
		return err
	}
	return nil
}

func (d *GeeRegistry) Discovery(updatePeers func(peers ...string)) {
	// 开启一个定时器,定时获取节点列表
	ticker := time.NewTicker(d.updateDuration)
	defer ticker.Stop()
	for range ticker.C {
		resp, err := http.Get(d.registry)
		if err != nil {
			log.Println("fail to get from registry:", err)
			continue
		}
		if resp.StatusCode != http.StatusOK {
			log.Println("fail to get from registry:", resp.StatusCode)
			continue
		}
		headerPeers := resp.Header.Get("X-Geerpc-Servers")
		if headerPeers != "" {
			peers := strings.Split(headerPeers, ",")
			// 比较节点列表是否发生变化
			if !reflect.DeepEqual(d.peers, peers) {
				d.peers = peers
				updatePeers(peers...)
			}
		}
		resp.Body.Close()
	}
}

func (d *GeeRegistry) Close() {

}

// Heartbeat 注册并发送心跳
func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		// 没有设定发送间隔的话
		// 默认比超时间隔早2s
		duration = defaultTimeout - time.Duration(2)*time.Second
	}
	var err error
	// 开启定时器前,先立即发一次,注册
	err = sendHeartbeat(registry, addr)
	// 开启一个新的协程,定时发送heartbeat
	go func() {
		t := time.NewTicker(duration)
		for err == nil {
			<-t.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-Geerpc-Server", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("fail to send heart beat to registry:", err)
		return err
	}
	return nil
}
