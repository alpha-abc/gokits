package packetv1_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/alpha-abc/gokits/packet/packetv1"
)

func Test_Test(t *testing.T) {
	// var a uint64 = 7
	// var u64 uint64 = (-1 ^ (-1 << (8 * a)))
	// fmt.Println(u64)
	var arr = make([]byte, 1<<54)
	var _ = arr[0 : 1<<62]

	var _ int = -1 ^ (-1 << (7 * 8))
}

func Test_EncodeUint64(t *testing.T) {
	fmt.Println(packetv1.EncodeUint64(4, 4278190081))
}

func Test_DecodeUint64(t *testing.T) {
	fmt.Println(packetv1.DecodeUint64([]byte{255, 255}))
}

func Test_PackUnPack(t *testing.T) {
	var p, _ = packetv1.NewPacket(1, 1, 1)
	var bs, _ = p.Pack(3, 4, make([]byte, 255))

	fmt.Println(bs)
	var buff bytes.Buffer
	for _, b := range bs {
		buff.Write(b)
	}

	var _v, _c, _d, _e = p.Unpack(&buff)
	fmt.Println(_v, _c, len(_d), _e)

	_v, _c, _d, _e = p.Unpack(&buff)
	fmt.Println(_v, _c, len(_d), _e)
}
