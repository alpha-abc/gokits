package packetv1

import (
	"bytes"
	"fmt"
	"io"
)

// https://www.sunbloger.com/2018/09/09/612.html

// EncodeUint64 将{u64}转换成长度为{n}的字节数组
func EncodeUint64(n int, u64 uint64) []byte {
	// 受限于uint64, 此处限制{n}长度范围: [0, 8]
	if n > 8 || n < 0 {
		panic("encode uint64 length error")
	}

	var bs = make([]byte, n)

	for i := 0; i < n; i++ {
		bs[i] = byte(u64 >> ((n - 1 - i) * 8))
	}

	return bs
}

// DecodeUint64 将字节数组(长度范围:[0, 8])转换成 uint64
func DecodeUint64(bs []byte) uint64 {
	var n = len(bs)
	// 受限于uint64, 此处限制{bs}长度范围: [0, 8]
	if n > 8 || n < 0 {
		panic("decode uint64 length error")
	}

	var u64 uint64 = 0
	for i := 0; i < n; i++ {
		u64 = u64 | uint64(bs[i])<<((n-1-i)*8)
	}

	return u64
}

// 包头格式
type Packet struct {
	vl int // version length, 版本号所占字节长度 [0,8]
	cl int // command length, 命令所占字节长度 [0,8]
	dl int // data length, 数据所占字节长度 [0,8]
}

// NewHeader 设置每个字段的长度
func NewPacket(vl int, cl int, ll int) (*Packet, error) {
	if vl < 0 || vl > 8 {
		return nil, fmt.Errorf("invalid vl")
	}
	if cl < 0 || cl > 8 {
		return nil, fmt.Errorf("invalid cl")
	}
	if ll < 0 || ll > 8 {
		return nil, fmt.Errorf("invalid ll")
	}

	return &Packet{
		vl: vl,
		cl: cl,
		dl: ll,
	}, nil
}

func (p *Packet) Pack(version uint64, command uint64, data []byte) ([][]byte, error) {
	var maxVer uint64
	if p.vl == 8 {
		maxVer = -1 ^ (-1 << (8 * 8))
	} else if p.vl < 8 {
		maxVer = uint64(int64(-1 ^ (-1 << (p.vl * 8))))
	}
	if version > maxVer {
		return nil, fmt.Errorf("invalid version, %d, max version, %d", version, maxVer)
	}

	var maxCmd uint64
	if p.cl == 8 {
		maxCmd = -1 ^ (-1 << (8 * 8))
	} else if p.cl < 8 {
		maxCmd = uint64(int64(-1 ^ (-1 << (p.cl * 8))))
	}
	if command > uint64(maxCmd) {
		return nil, fmt.Errorf("invalid command, %d, max command, %d", command, maxCmd)
	}

	var maxDataLen uint64
	if p.dl == 8 {
		maxDataLen = -1 ^ (-1 << (8 * 8))
	} else if p.dl < 8 {
		maxDataLen = uint64(int64(-1 ^ (-1 << (p.dl * 8))))
	}

	var dataLen = uint64(len(data))

	if maxDataLen == 0 || dataLen == 0 {
		var vbs = EncodeUint64(p.vl, version)
		var cbs = EncodeUint64(p.cl, command)
		var lbs = EncodeUint64(p.dl, 0)

		var bss [][]byte = make([][]byte, 1)
		var buff bytes.Buffer
		buff.Write(vbs)
		buff.Write(cbs)
		buff.Write(lbs)

		bss[0] = buff.Bytes()
		return bss, nil
	}

	var n = dataLen / maxDataLen
	if dataLen%maxDataLen > 0 {
		n += 1
	}

	var bss [][]byte = make([][]byte, n)

	for i := 0; i < int(n); i++ {
		var vbs = EncodeUint64(p.vl, version)
		var cbs = EncodeUint64(p.cl, command)

		var ds []byte
		if uint64(i) == n-1 {
			ds = data[uint64(i)*maxDataLen:]
		} else {
			ds = data[uint64(i)*maxDataLen : uint64(i+1)*maxDataLen]
		}
		var lbs = EncodeUint64(p.dl, uint64(len(ds)))

		var buff bytes.Buffer
		buff.Write(vbs)
		buff.Write(cbs)
		buff.Write(lbs)
		buff.Write(ds)

		bss[i] = buff.Bytes()
	}

	return bss, nil
}

func (p *Packet) Unpack(reader io.Reader) (uint64, uint64, []byte, error) {
	var headerBs = make([]byte, p.vl+p.cl+p.dl)

	var nr, er = reader.Read(headerBs)
	if nr != len(headerBs) {
		return 0, 0, nil, fmt.Errorf("unpack header error, read len: %d, need len: %d", nr, len(headerBs))
	}
	if er != nil {
		return 0, 0, nil, er
	}

	var i1 = 0
	var i2 = p.vl + i1
	var i3 = p.cl + i2
	var i4 = p.dl + i3

	var version = DecodeUint64(headerBs[i1:i2])
	var command = DecodeUint64(headerBs[i2:i3])
	var dl = DecodeUint64(headerBs[i3:i4])

	var dataBs = make([]byte, dl)
	nr, er = reader.Read(dataBs)
	if nr != len(dataBs) {
		return 0, 0, nil, fmt.Errorf("unpack data error, read len: %d, need len: %d", nr, len(headerBs))
	}
	if er != nil {
		if er != io.EOF {
			return 0, 0, nil, er
		}
	}

	return version, command, dataBs, nil
}
