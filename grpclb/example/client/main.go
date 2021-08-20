package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alpha-abc/gokits/grpclb/balancer"
	"github.com/alpha-abc/gokits/grpclb/example"
	"github.com/alpha-abc/gokits/grpclb/register"
	"github.com/alpha-abc/gokits/grpclb/resolver"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

var etcdAddr = "alpha.org:2379"

// 演示了“循环”和“hash”的负载均衡
func runWithEtcd3() {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 12 * time.Second,
	})

	if err != nil {
		panic(err)
	}
	// 后续hash传参, 用k定位
	var rrScheme = "resolver_test"
	var k interface{} = "default_key"
	balancer.RegistAddrHashPickerBuilder(balancer.AddrHashName, k.(string))
	resolver.RegistEtcd3Resolver(rrScheme, etcdCli, register.CreateEtcd3WatchKey("grpclb-etcd3", "test-server", "1.0.0"))

	var rrConn, rrErr = grpc.Dial(
		fmt.Sprintf("%s:///", rrScheme), // resolver 相关
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
	)
	if rrErr != nil {
		panic(rrErr)
	}

	var hashScheme = "resolver_test1"
	resolver.RegistEtcd3Resolver(hashScheme, etcdCli, register.CreateEtcd3WatchKey("grpclb-etcd3", "test-server", "1.0.0"))
	var hashConn, hashErr = grpc.Dial(
		fmt.Sprintf("%s:///", hashScheme), // resolver 相关
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, balancer.AddrHashName)), // picker 相关
	)
	if hashErr != nil {
		panic(rrErr)
	}

	defer rrConn.Close()
	defer hashConn.Close()

	var rrClient = example.NewGreeterClient(rrConn)
	var hashClient = example.NewGreeterClient(hashConn)
	for i := 0; i < 100; i++ {
		time.Sleep(1 * time.Second)

		if true {
			var res, err = rrClient.SayHello(context.Background(), &example.Request{Content: "round robin"})
			if err != nil {
				fmt.Println("err", err)
				return
			}
			fmt.Println(res)
		}

		if true {
			// 注意: 实际使用时, 如果没有对应的hash地址, 需要自己加超时处理
			var res1, err1 = hashClient.SayHello(context.WithValue(context.Background(), k, "127.0.0.1:8002"), &example.Request{Content: "addr hash"})
			if err1 != nil {
				fmt.Println("err1", err1)
				return
			}
			fmt.Println(res1)
		}
	}

}

func main() {
	runWithEtcd3()
}
