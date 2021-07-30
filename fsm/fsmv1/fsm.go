package fsmv1

import "fmt"

// 简易版状态机工具

var (
	// m is a map from scheme to resolver builder.
	m = make(map[Scheme]map[State]map[Event]Handler)
)

// Scheme 状态机
type Scheme string

// State 状态
type State string

// Event 事件
type Event string

// Handler 处理方法，返回新的状态
type Handler func(v ...interface{}) (State, error)

// Register 注册状态机
func Register(s Scheme) error {
	if _, ok := m[s]; ok {
		return fmt.Errorf("scheme (%s) has registed", s)
	}
	m[s] = make(map[State]map[Event]Handler)
	return nil
}

// AddHandler 添加处理事件
func AddHandler(s Scheme, state State, event Event, handler Handler) error {
	var m, ok = m[s]
	if !ok {
		return fmt.Errorf("scheme (%s) has not regist", s)
	}

	if _, ok := m[state]; !ok {
		m[state] = make(map[Event]Handler)
	}

	if _, ok := m[state][event]; ok {
		return fmt.Errorf("state(%s) event(%s) has defined", state, event)
	}

	m[state][event] = handler
	return nil
}

// Do 状态转换
// @parameters:
// 		@scheme 状态机名称
// 		@state 当前状态
//		@event 可处理事件
//      @v 自定义可变参数
// @return
//		@state 转换后的状态
func Do(scheme Scheme, state State, event Event, v ...interface{}) (State, error) {
	var m, ok = m[scheme]
	if !ok {
		return "", fmt.Errorf("scheme (%s) has not regist", scheme)
	}

	var eh, ok1 = m[state]
	if !ok1 {
		return "", fmt.Errorf("state(%s) has not define event and handler", state)
	}

	var fn, ok2 = eh[event]
	if !ok2 {
		return "", fmt.Errorf("event(%s) has no handlers", event)
	}

	return fn(v)
}
