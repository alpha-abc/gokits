package balancer

import (
	"context"
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

// 负载均衡，地址一致

// AddrHashName 默认名称
const AddrHashName = "addr_hash"

// RegistAddrHashPickerBuilder 初始化
// @name balancer.Builder名称，默认传递AddrHashName
// @addrHashKey 地址hash，通过context传递
func RegistAddrHashPickerBuilder(name, addrHashKey string) {
	balancer.Register(newAddrHashPickerBuilder(name, addrHashKey))
}

func newAddrHashPickerBuilder(name, addrHashKey string) balancer.Builder {
	return base.NewBalancerBuilderWithConfig(name, &addrHashPickerBuilder{addrHashKey: addrHashKey}, base.Config{HealthCheck: true})
}

type addrHashPickerBuilder struct {
	addrHashKey string
}

// Build implemet base.PickerBuilder
func (pb *addrHashPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var picker = &addrHashPicker{
		subConns:    make(map[string]balancer.SubConn),
		addrHashKey: pb.addrHashKey,
	}

	// 此处只是简单的添加地址作为key
	for addr, sc := range readySCs {
		picker.subConns[addr.Addr] = sc
	}

	return picker
}

type addrHashPicker struct {
	mux         sync.Mutex
	subConns    map[string]balancer.SubConn
	addrHashKey string
}

// Pick implement balancer.Picker
func (p *addrHashPicker) Pick(ctx context.Context, opts balancer.PickOptions) (balancer.SubConn, func(balancer.DoneInfo), error) {
	var sc balancer.SubConn
	p.mux.Lock()
	var key, ok = ctx.Value(p.addrHashKey).(string)
	if ok {
		var targetSubConn, found = p.subConns[key]
		if found {
			sc = targetSubConn
		}
	}
	p.mux.Unlock()
	return sc, nil, nil
}
