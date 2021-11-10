package init

import "sync"

// register

// run

type info struct {
	wg  sync.WaitGroup
	cnt int

	node *node
}

type node struct {
	fn   func()
	next *node
}

func New() *info {
	return &info{
		cnt:  0,
		node: nil,
	}
}

func (inf *info) Register(fn func()) {
	var header *node

	for n := inf.node; n != nil; n = n.next {
		header = n
	}

	var node = &node{
		fn:   fn,
		next: nil,
	}

	if header != nil {
		header.next = node
	} else {
		inf.node = node
	}

	inf.cnt++
}

func (inf *info) Do() {
	inf.wg.Add(inf.cnt)
	for n := inf.node; n != nil; n = n.next {
		go func() {
			defer inf.wg.Done()
			n.fn()
		}()
	}

	inf.wg.Wait()
}
