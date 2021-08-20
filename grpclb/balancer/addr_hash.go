package balancer

import (
	"sync"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

// 负载均衡，地址一致

// AddrHashName 默认名称
const AddrHashName = "addr_hash"

// RegistAddrHashPickerBuilder 初始化
// @name balancer.Builder名称，默认传递AddrHashName
// @addrHashKey 地址hash，通过context传递, {"<addrHashKey>": "<real hash addr>"}, 用于获取定位具体需要hash的地址
func RegistAddrHashPickerBuilder(name, addrHashKey string) {
	balancer.Register(newAddrHashPickerBuilder(name, addrHashKey))
}

func newAddrHashPickerBuilder(name, addrHashKey string) balancer.Builder {
	return base.NewBalancerBuilder(
		name,
		&addrHashPickerBuilder{
			addrHashKey: addrHashKey,
		},
		base.Config{
			HealthCheck: true,
		},
	)
	// return base.NewBalancerBuilderWithConfig(name, &addrHashPickerBuilder{addrHashKey: addrHashKey}, base.Config{HealthCheck: true})
}

type addrHashPickerBuilder struct {
	addrHashKey string
}

var _ base.PickerBuilder = new(addrHashPickerBuilder)

func (pb *addrHashPickerBuilder) Name() string {
	return AddrHashName
}

// Build implemet base.PickerBuilder
func (pb *addrHashPickerBuilder) Build(buildInfo base.PickerBuildInfo) balancer.Picker {
	if len(buildInfo.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var picker = &addrHashPicker{
		subConns:    make(map[string]balancer.SubConn),
		addrHashKey: pb.addrHashKey,
	}

	// 此处只是简单的添加地址作为key
	for sc, conInfo := range buildInfo.ReadySCs {
		picker.subConns[conInfo.Address.Addr] = sc
	}

	return picker
}

// func (pb *addrHashPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
// 	if len(readySCs) == 0 {
// 		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
// 	}

// 	var picker = &addrHashPicker{
// 		subConns:    make(map[string]balancer.SubConn),
// 		addrHashKey: pb.addrHashKey,
// 	}

// 	// 此处只是简单的添加地址作为key
// 	for addr, sc := range readySCs {
// 		picker.subConns[addr.Addr] = sc
// 	}

// 	return picker
// }

type addrHashPicker struct {
	mux         sync.Mutex
	subConns    map[string]balancer.SubConn
	addrHashKey string
}

var _ balancer.Picker = new(addrHashPicker)

// Pick implement balancer.Picker
func (p *addrHashPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	var pr balancer.PickResult

	p.mux.Lock()
	defer p.mux.Unlock()

	var key, ok = info.Ctx.Value(p.addrHashKey).(string)
	if !ok {
		return pr, nil
	}

	var targetSubConn, ok1 = p.subConns[key]
	if ok1 {
		pr.SubConn = targetSubConn
	}

	return pr, nil
}

// func (p *addrHashPicker) Pick(ctx context.Context, opts balancer.PickInfo) (balancer.SubConn, func(balancer.DoneInfo), error) {
// 	var sc balancer.SubConn
// 	p.mux.Lock()
// 	var key, ok = ctx.Value(p.addrHashKey).(string)
// 	if ok {
// 		var targetSubConn, found = p.subConns[key]
// 		if found {
// 			sc = targetSubConn
// 		}
// 	}
// 	p.mux.Unlock()
// 	return sc, nil, nil
// }
