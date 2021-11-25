package v0_test

import (
	"fmt"
	"testing"

	v0 "github.com/alpha-abc/gokits/httpclient/v0"
)

func Test_(t *testing.T) {
	var r = &v0.HTTPRequest{
		Method: "GET",
		URL:    "http://127.0.0.1:4242",
	}

	var res, err = r.Request()
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println(string(res.Body))
}
