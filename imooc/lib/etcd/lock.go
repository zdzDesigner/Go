package etcd

import (
	"context"
	"fmt"

	"github.com/coreos/etcd/clientv3"
)

// Lock ..
func Lock(key string, fn func()) {

	etcdClient := Client()
	// 租约
	lease := clientv3.NewLease(etcdClient)
	leaseResp, err := lease.Grant(context.TODO(), 100)
	if err != nil {
		return
	}
	leaseID := leaseResp.ID
	fmt.Println("leaseID:", leaseID)
	defer func() {
		fmt.Println("lease revoke")
		_, err := lease.Revoke(context.TODO(), leaseID)
		fmt.Println("lease revoke err:", err)
	}()

	// ctx, cancelFunc := context.WithCancel(context.TODO())
	// 预先释放
	// defer cancelFunc()

	// lkarCh, err := lease.KeepAlive(ctx, leaseID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// go func() {
	// 	for {
	// 		select {
	// 		case lkar := <-lkarCh:
	// 			if lkar == nil {
	// 				fmt.Println("租约失效")
	// 				return
	// 			}
	// 			fmt.Println("续租:", lkar.ID)
	// 		}
	// 	}
	// }()

	// 1. 创建锁(创建租约, 自动续租)
	kv := etcdClient.KV
	txn := kv.Txn(context.TODO())
	txn.If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
		Then(clientv3.OpPut(key, "exist", clientv3.WithLease(leaseID)))

	txnResp, err := txn.Commit()
	if err != nil {
		fmt.Println(err)
		return
	}
	if !txnResp.Succeeded {
		return
	}
	// 2. 处理业务
	fn()
	// 3. 释放锁 defer

}
