package snowflakev1

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func CreateNodeId(cli *clientv3.Client, key string, limit int64) (int64, error) {
	var resp, err = cli.Put(context.TODO(), key, "", clientv3.WithPrevKV())
	if err != nil {
		return 0, err
	}

	if resp.PrevKv == nil {
		return 1, nil
	}

	var ver = resp.PrevKv.Version + 1

	if ver%limit == 0 {
		cli.Delete(context.TODO(), key)
	}

	return resp.PrevKv.Version + 1, nil
}
