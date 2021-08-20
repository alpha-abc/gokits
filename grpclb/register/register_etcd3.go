package register

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const leaseIDZero clientv3.LeaseID = 0

// RegistEtcd3Registrar .
func RegistEtcd3Registrar(key, val string, etcd3Cli *clientv3.Client, interval, ttl int64) (*etcd3Registrar, error) {
	if len(key) == 0 {
		return nil, fmt.Errorf("invalid param: key")
	}

	if len(val) == 0 {
		return nil, fmt.Errorf("invalid param: val")
	}

	if etcd3Cli == nil {
		return nil, fmt.Errorf("invalid param: etcd3Cli")
	}

	if interval <= 0 {
		return nil, fmt.Errorf("invalid param: interval")
	}

	if ttl <= 0 {
		return nil, fmt.Errorf("invalid param: ttl")
	}

	if interval >= ttl-3 {
		return nil, fmt.Errorf("invalid interval and ttl")
	}

	var registrar = &etcd3Registrar{
		etcd3Cli: etcd3Cli,
		key:      key,
		val:      val,
		interval: interval,
		ttl:      ttl,
	}

	if err := registrar.regist(); err != nil {
		return nil, err
	}

	return registrar, nil
}

// etcd3Registrar .
type etcd3Registrar struct {
	key      string
	val      string
	interval int64 // second
	ttl      int64 // second

	etcd3Cli   *clientv3.Client
	leaseID    clientv3.LeaseID
	unRegistry chan int
}

// regist .
func (r *etcd3Registrar) regist() error {
	r.leaseID = leaseIDZero
	r.unRegistry = make(chan int, 0)

	go func() {
		var tiker = time.NewTicker(time.Duration(r.interval) * time.Second)

		for {
			var err = r.keepAlive()
			select {
			case <-tiker.C:
				if err != nil {
					// nothing to do
					continue
				}
			case <-r.unRegistry:
				r.etcd3Cli.Delete(context.Background(), r.key)
				return
			}
		}
	}()

	return nil
}

// UnRegist .
func (r *etcd3Registrar) UnRegist() {
	r.unRegistry <- 0
}

func (r *etcd3Registrar) keepAlive() error {
	var resGet, errGet = r.etcd3Cli.Get(context.Background(), r.key)
	if errGet != nil {
		r.leaseID = leaseIDZero
		return errGet
	}

	if resGet.Count == 0 || r.leaseID == 0 {
		resGrant, err := r.etcd3Cli.Grant(context.Background(), r.ttl)
		if err != nil {
			r.leaseID = leaseIDZero
			return err
		}

		r.leaseID = resGrant.ID
		if _, err := r.etcd3Cli.Put(context.Background(), r.key, r.val, clientv3.WithLease(r.leaseID)); err != nil {
			r.leaseID = leaseIDZero
			return err
		}
	}

	var _, errKA = r.etcd3Cli.KeepAliveOnce(context.Background(), r.leaseID)
	if errKA != nil {
		r.leaseID = leaseIDZero
		return errKA
	}

	return nil
}

// CreateEtcd3WatchKey .
func CreateEtcd3WatchKey(registryPrefix, name, version string) string {
	return fmt.Sprintf("/%s/%s/%s", registryPrefix, name, version)
}

// CreateEtcd3Key .
func CreateEtcd3Key(registryPrefix, name, version, node string) string {
	return fmt.Sprintf("%s/%s", CreateEtcd3WatchKey(registryPrefix, name, version), node)
}
