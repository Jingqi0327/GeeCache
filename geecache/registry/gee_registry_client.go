package registry

import (
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

type GeeRegistryClient struct {
	registry          string
	heartbeatDuration time.Duration
	updateDuration    time.Duration
	peers             []string
}

func NewGeeRegistryClient(registry string, heartbeatDuration time.Duration, updateDuration time.Duration) RegistryClient {
	return &GeeRegistryClient{
		registry:          registry,
		heartbeatDuration: heartbeatDuration,
		updateDuration:    updateDuration,
	}
}

func (d *GeeRegistryClient) Register(addr string, stop chan error) error {
	if d.heartbeatDuration == 0 {
		// 没有设定发送间隔的话,默认比超时间隔早2s
		d.heartbeatDuration = defaultTimeout - time.Duration(2)*time.Second
	}
	ticker := time.NewTicker(d.heartbeatDuration)
	defer ticker.Stop()

	// 先立即发送一次,注册
	err := sendHeartbeat(d.registry, addr)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			// 定时发送心跳
			err = sendHeartbeat(d.registry, addr)
			if err != nil {
				return err
			}
		case err := <-stop:
			// 停止发送心跳
			return err
		}
	}
}

func sendHeartbeat(registry, addr string) error {
	//log.Println(addr, "send heart beat to registry", registry)
	httpClient := &http.Client{}
	req, _ := http.NewRequest("POST", registry, nil)
	req.Header.Set("X-GeeRegistry-Servers", addr)
	if _, err := httpClient.Do(req); err != nil {
		log.Println("fail to send heart beat to registry:", err)
		return err
	}
	return nil
}

func (d *GeeRegistryClient) Discovery(updatePeers func(peers ...string)) {
	// 获取节点列表
	fetchPeers := func() {
		resp, err := http.Get(d.registry)
		if err != nil {
			log.Println("fail to get from registry:", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Println("fail to get from registry, status:", resp.StatusCode)
			return
		}
		headerPeers := resp.Header.Get("X-GeeRegistry-Servers")
		if headerPeers != "" {
			peers := strings.Split(headerPeers, ",")
			if !reflect.DeepEqual(d.peers, peers) {
				d.peers = peers
				updatePeers(peers...)
			}
		}
	}

	// 先获取一次节点列表
	fetchPeers()

	// 开启一个定时器,定时获取节点列表
	ticker := time.NewTicker(d.updateDuration)
	defer ticker.Stop()
	for range ticker.C {
		fetchPeers()
	}
}

func (d *GeeRegistryClient) Close() {

}
