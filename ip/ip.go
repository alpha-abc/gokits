package ip

import (
	"net"
	"regexp"
	"strings"
)

// GetLocalIPV4 获取检测到的第一个IPV4地址，本机地址除外
func GetLocalIPV4() string {
	var localIP = ""

	var addrs, errIA = net.InterfaceAddrs()
	if errIA != nil {
		return localIP
	}

	var reg, err = regexp.Compile(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/[0-9]+$`)
	if err != nil {
		return localIP
	}

	for _, addr := range addrs {
		var match = reg.MatchString(addr.String())
		// var match, _ = regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+/[0-9]+$`, addr.String())
		if match {
			if strings.HasPrefix(addr.String(), "127.0.0.1") {
				continue
			}
			localIP = strings.Split(addr.String(), "/")[0]
			break
		}
	}

	return localIP
}
