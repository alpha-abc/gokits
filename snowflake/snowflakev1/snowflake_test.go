package snowflakev1_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alpha-abc/gokits/snowflake/snowflakev1"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Test(t *testing.T) {
	var node, e = snowflakev1.NewNode(2)
	if e != nil {
		fmt.Println(e.Error())
		return
	}

	for i := 0; i < 100; i++ {
		var id, err = node.Generate()
		fmt.Println(id, err)
		var head, tt, node, seq = snowflakev1.Extract(id)
		fmt.Println(head, tt, node, seq)
	}
}

func Test_NodeID(t *testing.T) {
	cli, e := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 12 * time.Second,
	})

	if e != nil {
		println(e)
		return
	}

	var n = 5
	var wg sync.WaitGroup
	wg.Add(5)

	for i := 0; i < n; i++ {
		go func() {
			var id, err = snowflakev1.CreateNodeId(cli, "/k", 32)
			if err != nil {
				println(err)
			} else {
				println(id)
			}

			wg.Done()
		}()
	}

	wg.Wait()
}
