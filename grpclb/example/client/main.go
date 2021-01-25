package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alpha-abc/gokits/grpclb/balancer"
	"github.com/alpha-abc/gokits/grpclb/example"
	"github.com/alpha-abc/gokits/grpclb/register"
	"github.com/alpha-abc/gokits/grpclb/resolver"
	"github.com/coreos/etcd/clientv3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
)

func runWithEtcd3() {
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 12 * time.Second,
	})

	if err != nil {
		panic(err)
	}
	var k interface{} = "default_key"
	balancer.RegistAddrHashPickerBuilder(balancer.AddrHashName, k.(string))

	resolver.RegistEtcd3Resolver("resolver_test", etcdCli, register.CreateEtcd3WatchKey("grpclb-etcd3", "test-server", "1.0.0"))

	var conn, errD = grpc.Dial(
		"resolver_test:///", // resolver 相关
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)),
	)
	if errD != nil {
		panic(errD)
	}

	resolver.RegistEtcd3Resolver("resolver_test1", etcdCli, register.CreateEtcd3WatchKey("grpclb-etcd3", "test-server", "1.0.0"))
	var conn1, errD1 = grpc.Dial(
		"resolver_test1:///", // resolver 相关
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, balancer.AddrHashName)), // picker 相关
	)
	if errD1 != nil {
		panic(errD)
	}

	defer conn.Close()
	defer conn1.Close()

	var client = example.NewGreeterClient(conn)
	var client1 = example.NewGreeterClient(conn1)
	for i := 0; i < 300000; i++ {
		time.Sleep(1 * time.Second)
		var res, err = client.SayHello(context.Background(), &example.Request{Content: "round robin"})
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(res)

		var res1, err1 = client1.SayHello(context.WithValue(context.Background(), k, "127.0.0.1:8001"), &example.Request{Content: "addr hash"})
		if err1 != nil {
			fmt.Println(err1)
			return
		}
		fmt.Println(res1)
	}

}

func main() {
	runWithEtcd3()
}
