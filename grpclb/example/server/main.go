package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/alpha-abc/gokits/grpclb/example"
	"github.com/alpha-abc/gokits/grpclb/register"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

var port = flag.String("port", "8001", "listening port")
var ip = "127.0.0.1"

var etcdAddr = "alpha.org:2379"

type rpcServer struct {
	ip   string
	port string
}

func (r *rpcServer) SayHello(ctx context.Context, req *example.Request) (*example.Response, error) {
	fmt.Println(r.ip, r.port, "from client", req.Content)
	return &example.Response{
		Content: fmt.Sprintf("reply %s", req.Content),
	}, nil
}

func runWithEtcd3() {
	var port = *port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		panic(err)
	}

	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 12 * time.Second,
	})

	if err != nil {
		panic(err)
	}

	var key = register.CreateEtcd3Key("grpclb-etcd3", "test-server", "1.0.0", fmt.Sprintf("%s:%s", ip, port))
	fmt.Println("key", key)
	var val = fmt.Sprintf("%s:%s", ip, port)

	if _, err := register.RegistEtcd3Registrar(key, val, etcdCli, 5, 9); err != nil {
		panic(err)
	}

	var s = grpc.NewServer()
	example.RegisterGreeterServer(s, &rpcServer{
		ip:   ip,
		port: port,
	})

	fmt.Println("run server:", ip, port)
	err = s.Serve(lis)
	if err != nil {
		panic(err)
	}
}

// go run example/server/main.go -port 8001
// go run example/server/main.go -port 8002
// go run example/server/main.go -port 8003
func main() {
	flag.Parse()
	runWithEtcd3()
}
