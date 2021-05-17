package snowflakev1

import (
	"fmt"
	"sync"
	"time"
)

/*
[head: 1bit]   [         timestamp millisecond: 41bit            ]   [node:10bit]   [ seq: 12bit ]
    0    -     00000000  00000000  00000000  00000000  00000000  0 - 00000000  00 - 00000000  0000
*/

var (
	// Epoch 定义一个计算的起始时间, 毫秒时间戳
	Epoch int64 = 1563166498000

	// head占用比特位数
	headBitLen uint8 = 1

	// time占用比特位数 (相对Epoch存储的时间戳)
	timeBitLen uint8 = 41

	// node占用比特位数
	nodeBitLen uint8 = 10

	// sequence占用比特位数
	sequenceBitLen uint8 = 10
)

// ID snowflake id
type ID int64

// Node 生成ID的基本节点
type Node struct {
	mux sync.Mutex

	epoch          int64
	headBitLen     uint8
	timeBitLen     uint8
	nodeBitLen     uint8
	sequenceBitLen uint8

	relativeTimestamp int64
	nodeID            int64
	sequence          int64
}

func currMillisecond() int64 {
	return time.Now().UnixNano() / 1e6
}

// NewNode 初始化节点
func NewNode(nodeID int64) (*Node, error) {
	var maxNodeID = -1 ^ (-1 << nodeBitLen) // bit位数为{nodeBitLen}时表示的最大数
	if nodeID < 0 || nodeID > int64(maxNodeID) {
		return nil, fmt.Errorf("invalid noedeID, require [%d - %d]", 0, maxNodeID)
	}

	var now = currMillisecond()
	if now < Epoch {
		return nil, fmt.Errorf("epoch: %d greater than current millisecond: %d", Epoch, now)
	}

	return &Node{
		epoch:          Epoch,
		headBitLen:     headBitLen,
		timeBitLen:     timeBitLen,
		nodeBitLen:     nodeBitLen,
		sequenceBitLen: sequenceBitLen,

		relativeTimestamp: now - Epoch,
		nodeID:            nodeID,
		sequence:          0,
	}, nil

}

// Generate 返回一个唯一的ID
// 确保以下两点:
// - 保证系统时间的精确性
// - 保证任意两个Node不能有相同的节点ID
func (n *Node) Generate() (ID, error) {
	n.mux.Lock()
	defer n.mux.Unlock()

	var now = currMillisecond() // 当前毫秒时间戳
	var nodeTime = n.relativeTimestamp + n.epoch

	if now < nodeTime {
		return 0, fmt.Errorf("clock(%d) move backwards, reject generate until %d", now, nodeTime)
	}

	if now == nodeTime {
		n.sequence = (n.sequence + 1) & (-1 ^ (-1 << n.sequenceBitLen))

		if n.sequence == 0 {
			// 等待下一毫秒
			for now <= nodeTime {
				now = currMillisecond()
			}
		}
	} else {
		n.sequence = 0
	}

	n.relativeTimestamp = now - n.epoch

	var resp = (n.relativeTimestamp << (n.nodeBitLen + n.sequenceBitLen)) |
		(n.nodeID << n.sequenceBitLen) |
		n.sequence

	return ID(resp), nil
}

// Extract 根据ID提取head,当时时间戳, 节点ID, 序列号
func Extract(id ID) (int64, int64, int64, int64) {
	var i = int64(id)

	var head = i >> (timeBitLen + nodeBitLen + sequenceBitLen)
	var timestamp = (i >> (nodeBitLen + sequenceBitLen)) - ((i >> (timeBitLen + nodeBitLen + sequenceBitLen)) << timeBitLen) + Epoch
	var node = (i >> sequenceBitLen) - ((i >> (nodeBitLen + sequenceBitLen)) << nodeBitLen)
	var sequence = i - ((i >> sequenceBitLen) << sequenceBitLen)

	return head, timestamp, node, sequence
}
