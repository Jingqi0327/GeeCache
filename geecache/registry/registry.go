package registry

import (
	"context"
	"log"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdRegisrty struct {
	client      *clientv3.Client
	serviceName string
}

func NewEtcdRegisrty(endpoints []string, serviceName string) (*EtcdRegisrty, error) {
	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}
	client, err := clientv3.New(cfg)
	if err != nil {
		return nil, err
	}
	return &EtcdRegisrty{
		client:      client,
		serviceName: serviceName,
	}, nil
}

func (r *EtcdRegisrty) Close() {
	if r.client != nil {
		r.client.Close()
	}
}

func (r *EtcdRegisrty) Register(addr string, stop chan error) error {
	rsp, err := r.client.Grant(context.Background(), 5)
	if err != nil {
		return err
	}

	key := r.serviceName + addr
	_, err = r.client.Put(context.Background(), key, addr, clientv3.WithLease(rsp.ID))
	if err != nil {
		return err
	}

	keepAliveCh, err := r.client.KeepAlive(context.Background(), rsp.ID)
	if err != nil {
		return err
	}

	log.Printf("[Registry] Node %s register successful", addr)

	for {
		select {
		case err := <-stop:
			r.client.Delete(context.Background(), key)
			log.Printf("[Registry] Node %s logout", addr)
			return err
		case <-r.client.Ctx().Done():
			return nil
		case _, ok := <-keepAliveCh:
			if !ok {
				log.Printf("[Registry] Node %s timeout", addr)
				return nil
			}
		}
	}
}

func (r *EtcdRegisrty) Discovery(updatePeers func(peers ...string)) {
	rsp, err := r.client.Get(context.Background(), r.serviceName, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	var peers []string
	for _, kv := range rsp.Kvs {
		peers = append(peers, string(kv.Value))
	}
	updatePeers(peers...)

	watchCh := r.client.Watch(context.Background(), r.serviceName, clientv3.WithPrefix())
	for watchRsp := range watchCh {
		for _, ev := range watchRsp.Events {
			addr := string(ev.Kv.Value)
			switch ev.Type {
			case clientv3.EventTypePut:
				if !contains(peers, addr) {
					peers = append(peers, addr)
					updatePeers(peers...)
				}
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
