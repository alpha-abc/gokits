package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// JSONBody http body返回模型
type JSONBody struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// JComplete 将HTTP结果格式化成json，写回客户端
// @statusCode: httpcode
// @bizCode: 业务编码
// @message: 消息提示(业务消息或者http默认消息)
// @data: 数据
var JComplete = func(w http.ResponseWriter, statusCode int, bizCode int, message string, data interface{}) {
	if message == "" {
		message = http.StatusText(statusCode)
		// 若message还是为空，表示http状态码不正确
	}

	var resp = &JSONBody{
		Code:    fmt.Sprintf("%d.%d", statusCode, bizCode),
		Message: message,
		Data:    data,
	}

	/*同时设置header和http code，先设置header，然后code，否则header不生效*/
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	var buffer = bytes.NewBufferString("")
	var enc = json.NewEncoder(buffer)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	var _ = /*error ignore*/ enc.Encode(resp)

	//var b, _ = json.Marshal(res)
	fmt.Fprint(w, buffer.String())
	return
}

// 默认业务参数设置
const bizCode = 0

// 默认消息提醒设置
const message = ""

// JOK 正常返回，无数据实体
var JOK = func(w http.ResponseWriter) {
	JComplete(w, http.StatusOK, bizCode, message, nil)
}

// JResult 正常返回，有数据实体
var JResult = func(w http.ResponseWriter, v interface{}) {
	JComplete(w, http.StatusOK, bizCode, message, v)
}

// JReject 非正常返回，无数据实体
// @statusCode 参考标准库参数 http.Status*
var JReject = func(w http.ResponseWriter, statusCode, bizCode int, message string) {
	JComplete(w, statusCode, bizCode, message, nil)
}

// routeRecover 捕获router上的panic
var routeRecover = func(w http.ResponseWriter, r *http.Request) {
	var v = recover()
	if v != nil {
		JReject(w, http.StatusInternalServerError, bizCode, fmt.Sprint(v))
	}
}

// DecoratorFunc 路由装饰
// @f 原路由函数
// @fns 验证函数
// @return 装饰后的新函数
func DecoratorFunc(
	f func(http.ResponseWriter, *http.Request),
	fns ...func(*http.Request) error,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer routeRecover(w, r)

		for _, fn := range fns {
			var err = fn(r)
			if err != nil {
				JReject(w, http.StatusForbidden, bizCode, err.Error())
				return
			}
		}

		f(w, r)
		return
	}
}

// JRouteReject for rest json
type JRouteReject struct {
	StatusCode int    // http code
	StatusText string // http message
}

func (rr *JRouteReject) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	JReject(w, rr.StatusCode, bizCode, rr.StatusText)
}

// NotFound 自定义404格式
var NotFound = JRouteReject{
	StatusCode: http.StatusNotFound,
	StatusText: http.StatusText(http.StatusNotFound),
}

// MethodNotAllowed 自定义405格式
var MethodNotAllowed = JRouteReject{
	StatusCode: http.StatusMethodNotAllowed,
	StatusText: http.StatusText(http.StatusMethodNotAllowed),
}
