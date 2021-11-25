package v0

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HTTPRequest 请求结构体
type HTTPRequest struct {
	Method        string
	URL           string
	Params        map[string]string // url参数
	Headers       map[string]string
	Body          string            // 其他 Content-Type 类型
	Form          map[string]string // 只适用 application/x-www-form-urlencoded
	MultipartForm map[string]string // 只适用 multipart/form-data
	Timeout       int               // 单位 second
}

// HTTPResponse 响应结构体
type HTTPResponse struct {
	Proto      string // e.g. "HTTP/1.0"
	StatusCode int    // e.g. 200
	Status     string // e.g. "200 OK"
	Headers    map[string][]string
	Body       []byte
}

// Request http client
func (req *HTTPRequest) Request() (*HTTPResponse, error) {
	if strings.Trim(req.Method, " ") == "" {
		return nil, errors.New("invalid http method")
	}

	if strings.Trim(req.URL, " ") == "" {
		return nil, errors.New("invalid http url")
	}

	if req.Timeout == 0 {
		req.Timeout = 10
	}

	var u, uErr = url.Parse(req.URL)
	if uErr != nil {
		return nil, uErr
	}

	var pms = u.Query()
	for k, v := range req.Params {
		pms.Add(k, v)
	}

	var urlBuffer bytes.Buffer
	urlBuffer.WriteString(u.Scheme)
	urlBuffer.WriteString("://")
	urlBuffer.WriteString(u.Host)
	urlBuffer.WriteString(u.Path)

	if len(pms) > 0 {
		urlBuffer.WriteString("?")
		urlBuffer.WriteString(pms.Encode())
	}

	var body io.Reader
	switch req.Headers["Content-Type"] {
	case "multipart/form-data":
		var multipartBody bytes.Buffer
		var writer = multipart.NewWriter(&multipartBody)

		var wErr = func() error {
			defer func() { var _ = writer.Close() }()

			for k, v := range req.MultipartForm {
				if strings.HasPrefix(v, "file://") {
					var filePath = v[7:]

					var file, fErr = os.Open(filePath)
					if fErr != nil {
						return fErr
					}
					defer func() { var _ = file.Close() }()

					var part, partErr = writer.CreateFormFile(k, filepath.Base(filePath))
					if partErr != nil {
						return partErr
					}

					var _, copyErr = io.Copy(part, file)
					if copyErr != nil {
						return copyErr
					}
					continue
				}

				var wfErr = writer.WriteField(k, v)
				if wfErr != nil {
					return wfErr
				}
			}
			return nil
		}()
		if wErr != nil {
			return nil, wErr
		}

		req.Headers["Content-Type"] = writer.FormDataContentType()
		body = &multipartBody
	case "application/x-www-form-urlencoded":
		var s strings.Builder
		for k, v := range req.Form {
			if s.Len() > 0 {
				s.WriteString("&")
			}
			s.WriteString(url.QueryEscape(k))
			s.WriteString("=")
			s.WriteString(url.QueryEscape(v))
		}
		body = strings.NewReader(s.String())
	case "application/json":
		body = strings.NewReader(req.Body)
	default:
		body = strings.NewReader(req.Body)
	}

	var r, rErr = http.NewRequest(strings.ToUpper(req.Method), urlBuffer.String(), body)
	if rErr != nil {
		return nil, rErr
	}

	for k, v := range req.Headers {
		r.Header.Set(k, v)
	}

	var transport = &http.Transport{
		DisableKeepAlives: true,
	}

	var client = &http.Client{
		Transport: transport,
		Timeout:   time.Duration(req.Timeout) * time.Second,
	}

	var res, resErr = client.Do(r)
	if resErr != nil {
		return nil, resErr
	}
	defer res.Body.Close()

	var resBody, rbErr = ioutil.ReadAll(res.Body)
	if rbErr != nil {
		return nil, rbErr
	}

	return &HTTPResponse{
		Proto:      res.Proto,
		StatusCode: res.StatusCode,
		Status:     res.Status,
		Headers:    res.Header,
		Body:       resBody,
	}, nil
}
