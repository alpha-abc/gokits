package errorsv1

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

const (
	defaultNilText string = "<nil>"
)

// Error 自定义error
type Error struct {
	Code    string
	Desc    string
	WrapErr error
	Stack   string
}

func (e *Error) Error() string {
	if e == nil {
		return defaultNilText
	}

	return fmt.Sprintf("[%s] %s", e.Code, e.Desc)
}

// DetailString 输出错误详细信息
func DetailString(e error) string {
	var buf bytes.Buffer
	var err = e

	if err == nil {
		return defaultNilText
	}

	var idx = 1
	for err != nil {
		if pErr, ok := err.(*Error); ok {
			buf.WriteString(fmt.Sprintf("%d %s:\n    %s\n", idx, pErr.Stack, pErr.Error()))
		} else {
			buf.WriteString(fmt.Sprintf("%d ****:\n    %s\n", idx, err.Error()))
		}

		err = Unwrap(err)
		idx++
	}

	return buf.String()
}

// New 初始化, 并附带stack信息
func New(code, desc string) error {
	return &Error{
		Code:    code,
		Desc:    desc,
		WrapErr: nil,
		Stack:   getStack(2),
	}
}

// Wrap 包装一个新的错误信息
func Wrap(code, desc string, err error) error {
	return &Error{
		Code:    code,
		Desc:    desc,
		WrapErr: err,
		Stack:   getStack(2),
	}
}

// Unwrap 解包返回内层错误信息
func Unwrap(err error) error {
	if e, ok := err.(*Error); ok {
		return e.WrapErr
	}

	return errors.Unwrap(err)
}

// 获取运行时所需信息
func getStack(skip int) string {
	var pc, fileName, lineNumber, ok = runtime.Caller(skip)
	var funcName = ""
	if !ok {
		return "runtime caller failure"
	}

	for i := len(fileName) - 1; i > 0; i-- {
		if fileName[i] == '/' {
			fileName = fileName[i+1:]
			break
		}
	}

	var buf bytes.Buffer

	funcName = runtime.FuncForPC(pc).Name()

	buf.WriteString("(")
	buf.WriteString(funcName)
	buf.WriteString(") ")
	buf.WriteString(fileName)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(lineNumber))

	return buf.String()
}
