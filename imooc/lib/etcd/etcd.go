package etcd

import (
	"fmt"
	"sync"
	"time"

	"github.com/coreos/etcd/clientv3"
)

var etcdClient *clientv3.Client
var once sync.Once

// Client etcd client
func Client() *clientv3.Client {
	once.Do(func() {
		etcdClient = conn()
	})
	return etcdClient
}

// conn ..
func conn() *clientv3.Client {

	config := clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	}

	client, err := clientv3.New(config)
	fmt.Println("etcd client:", client)
	if err != nil {
		panic(fmt.Sprintf("etcd error:%s", err))
	}
	return client
}
