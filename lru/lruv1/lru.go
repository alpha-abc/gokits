package lruv1

import (
	"bytes"
	"container/list"
	"fmt"
)

// LRU - Least Recently Used - 最近最少使用
// 相对于仅考虑时间因素的FIFO(First In First Out - 先进先出)
// 和仅考虑访问频率的LFU(Least Frequently Used - 最少使用),
// LRU 算法可以认为是相对平衡的一种淘汰算法.

// LRU定义: 如果数据最近被访问过, 那么将来被访问的概率也会更高.
// 核心数据结构: 字典和链表
// 解决问题点: 当处于满内存状态时, 如何存储新值, 淘汰旧值

// 具体操作:
// 1 新数据插入到链表头部.
// 2 每当缓存命中(即缓存数据被访问), 则将数据移到链表头部.
// 3 当链表满的时候, 将链表尾部的数据丢弃.

// DeleteFunc 触发函数, 可选, 删除时触发
type DeleteFunc func(string, Value)

// Value 用于计算占用内存
type Value interface {
	Len() int
}

// Cache LRU缓存, 并发不安全
type Cache struct {
	// 只能粗略限制计算内存大小
	maxBytes  int64 // 允许使用的最大内存
	currBytes int64 // 当前使用的内存

	keysMap  map[string]*list.Element // key字典
	linkList *list.List               // value链表

	onDelete DeleteFunc
}

type entity struct {
	key string
	val Value
}

func (e *entity) Len() int {
	return len(e.key) + e.val.Len()
}

// New 初始化
func New(maxBytes int64, onDelete DeleteFunc) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		currBytes: 0,
		keysMap:   make(map[string]*list.Element),
		linkList:  list.New(),
		onDelete:  onDelete,
	}
}

// Get 根据key获取元素
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.keysMap[key]; ok {
		c.linkList.MoveToFront(ele)

		var e = ele.Value.(*entity)
		return e.val, true
	}

	return nil, false
}

// Del 根据key删除元素, 若key为空, 则根据LRU规则删除
func (c *Cache) Del(key string) bool {
	// 移除最近最少访问节点, 即队尾元素
	if len(key) == 0 {
		if ele := c.linkList.Back(); ele != nil {
			c.linkList.Remove(ele)

			var e = ele.Value.(*entity)
			delete(c.keysMap, e.key)

			c.currBytes -= int64(e.Len())

			if c.onDelete != nil {
				c.onDelete(e.key, e.val)
			}

			return true
		}

		return false
	}

	// 移除指定元素
	if ele, ok := c.keysMap[key]; ok {
		c.linkList.Remove(ele)
		var e = ele.Value.(*entity)
		delete(c.keysMap, e.key)

		c.currBytes -= int64(e.Len())

		if c.onDelete != nil {
			c.onDelete(e.key, e.val)
		}

		return true
	}

	return false
}

// Add 添加元素
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.keysMap[key]; ok {
		c.linkList.MoveToFront(ele)

		var e = ele.Value.(*entity)
		c.currBytes += int64(value.Len() - e.val.Len())

		e.val = value
	} else {
		var e = &entity{
			key: key,
			val: value,
		}
		var ele = c.linkList.PushFront(e)

		c.keysMap[key] = ele
		c.currBytes += int64(e.Len())
	}

	for c.maxBytes != 0 && c.maxBytes < c.currBytes {
		c.Del("")
	}
}

func (c *Cache) String() string {
	var buf bytes.Buffer

	buf.WriteString(fmt.Sprintf("maxBytes: %d, currBytes: %d\n", c.maxBytes, c.currBytes))

	buf.WriteString("keysMap:{ ")
	for key := range c.keysMap {
		buf.WriteString(key)
		buf.WriteString(" ")
	}
	buf.WriteString("}\n")

	buf.WriteString("linkList:\n")
	for ele := c.linkList.Front(); ele != nil; ele = ele.Next() {
		var e = ele.Value.(*entity)
		buf.WriteString(fmt.Sprintf("key: %s, value: %v\n", e.key, e.val))
	}
	buf.WriteString("\n")

	return buf.String()
}
