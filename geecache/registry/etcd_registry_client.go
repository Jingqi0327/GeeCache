package registry

import (
	"context"
	"log"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdRegistry 是 Etcd 注册中心
type EtcdRegistryClient struct {
	client      *clientv3.Client
	serviceName string
}

// NewEtcdRegistry 创建新的 Etcd 注册中心
func NewEtcdRegistryClient(endpoints []string, serviceName string) (RegistryClient, error) {
	// 配置 etcd 客户端
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	// 创建 etcd 客户端
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &EtcdRegistryClient{
		client:      client,
		serviceName: serviceName,
	}, nil
}

// Close 关闭 Etcd 注册中心
func (r *EtcdRegistryClient) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

// Register 注册节点
func (r *EtcdRegistryClient) Register(addr string, stop chan error) error {
	// 申请一个租约，租约时间为 5 秒
	rsp, err := r.client.Grant(context.Background(), 5)
	if err != nil {
		return err
	}

	// 注册节点
	// 将节点信息存储到 etcd 中，并与租约绑定
	key := r.serviceName + addr
	_, err = r.client.Put(context.Background(), key, addr, clientv3.WithLease(rsp.ID))
	if err != nil {
		return err
	}

	// 保持租约激活
	// KeepAlive 会在后台开一个 goroutine，定期向 etcd 发送心跳，保持租约激活
	keepAliveCh, err := r.client.KeepAlive(context.Background(), rsp.ID)
	if err != nil {
		return err
	}

	log.Printf("[Registry] Node %s register successful", addr)

	// 监听各个退出的信号
	for {
		select {
		// 收到 stop 信号，注销节点
		case err := <-stop:
			r.client.Delete(context.Background(), key)
			log.Printf("[Registry] Node %s logout", addr)
			return err
		// etcd 客户端关闭
		case <-r.client.Ctx().Done():
			return nil
		// 节点超时未发送心跳
		case _, ok := <-keepAliveCh:
			if !ok {
				log.Printf("[Registry] Node %s timeout", addr)
				return nil
			}
		}
	}
}

// Discovery 发现节点
func (r *EtcdRegistryClient) Discovery(updatePeers func(peers ...string)) {
	// 获取所有节点
	rsp, err := r.client.Get(context.Background(), r.serviceName, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	var peers []string
	for _, kv := range rsp.Kvs {
		peers = append(peers, string(kv.Value))
	}
	// 更新节点列表
	updatePeers(peers...)

	// 监听节点变化
	watchCh := r.client.Watch(context.Background(), r.serviceName, clientv3.WithPrefix())
	// 阻塞等待节点变化
	for watchRsp := range watchCh {
		for _, ev := range watchRsp.Events {
			addr := string(ev.Kv.Value)
			switch ev.Type {
			// 添加节点
			case clientv3.EventTypePut:
				if !contains(peers, addr) {
					peers = append(peers, addr)
					updatePeers(peers...)
				}
			// 删除节点
			case clientv3.EventTypeDelete:
				keyStr := string(ev.Kv.Key)
				delAddr := strings.TrimPrefix(keyStr, r.serviceName)
				peers = remove(peers, delAddr)
				updatePeers(peers...)
			}
		}
	}
}

func contains(peers []string, addr string) bool {
	for _, v := range peers {
		if v == addr {
			return true
		}
	}
	return false
}

func remove(peers []string, addr string) []string {
	var res []string
	for _, v := range peers {
		if v == addr {
			continue
		}
		res = append(res, v)
	}
	return res
}
