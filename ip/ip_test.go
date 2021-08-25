package ip_test

import (
	"fmt"
	"testing"

	"github.com/alpha-abc/gokits/ip"
)

func Test_LocalIP(t *testing.T) {
	var s = ip.GetLocalIPV4()
	fmt.Println(s)
}
