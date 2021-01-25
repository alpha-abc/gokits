package resolver

import (
	"context"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"google.golang.org/grpc/resolver"
)

type etcd3Resolver struct {
	scheme    string // resolver全局名称
	etcd3Cli  *clientv3.Client
	watchPath string

	addrDict map[string]resolver.Address

	cc  resolver.ClientConn
	mux sync.Mutex
}

// RegistEtcd3Resolver 注册resolver
func RegistEtcd3Resolver(scheme string, etcd3Cli *clientv3.Client, watchPath string) {
	resolver.Register(&etcd3Resolver{
		scheme:    scheme,
		etcd3Cli:  etcd3Cli,
		watchPath: watchPath,
	})
}

// Build implement resolver.Builder
func (r *etcd3Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOption) (resolver.Resolver, error) {
	r.cc = cc
	r.addrDict = make(map[string]resolver.Address)

	go r.watch()

	return r, nil
}

// Scheme implement resolver.Builder
func (r *etcd3Resolver) Scheme() string {
	return r.scheme
}

// Close implement resolver.Resolver
func (r *etcd3Resolver) Close() {
}

// ResolveNow implement resolver.Resolver
func (r *etcd3Resolver) ResolveNow(rn resolver.ResolveNowOption) {

}

func (r *etcd3Resolver) watch() {
	var resGet, errGet = r.etcd3Cli.Get(context.Background(), r.watchPath, clientv3.WithPrefix())
	if errGet == nil {
		for _, item := range resGet.Kvs {
			r.addrDict[string(item.Key)] = resolver.Address{
				Addr: string(item.Value),
			}
		}
		r.update()
	}

	var chn = r.etcd3Cli.Watch(context.Background(), r.watchPath, clientv3.WithPrefix())
	for n := range chn {
		for _, ev := range n.Events {
			switch ev.Type {
			case mvccpb.PUT:
				r.addrDict[string(ev.Kv.Key)] = resolver.Address{
					Addr: string(ev.Kv.Value),
				}
			case mvccpb.DELETE:
				delete(r.addrDict, string(ev.Kv.Key))
			}
		}
		r.update()
	}
}

func (r *etcd3Resolver) update() {
	var lst []resolver.Address
	for _, v := range r.addrDict {
		lst = append(lst, v)
	}

	r.cc.UpdateState(resolver.State{
		Addresses: lst,
	})
}
