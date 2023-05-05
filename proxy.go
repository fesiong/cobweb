package cobweb

import (
	"log"
	"net"
	"net/url"
	"time"
)

var ProxyValid = false

func init() {
	for {
		checkProxy()
		time.Sleep(10 * time.Second)
	}
}

func checkProxy() {
	// 检查proxy ip
	if JsonData.DBConfig.Proxy == "" {
		ProxyValid = false
		return
	}
	proxyUrl, err := url.Parse(JsonData.DBConfig.Proxy)
	if err == nil {
		addr := net.JoinHostPort(proxyUrl.Hostname(), proxyUrl.Port())
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err != nil {
			ProxyValid = false
		} else {
			log.Println("proxy valid")
			ProxyValid = true
			conn.Close()
		}
	}
}
