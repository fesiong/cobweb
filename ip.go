package cobweb

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"time"
)

// 通过IP ping端口来查找

func Ping(ip, port string) bool {
	addr := net.JoinHostPort(ip, port)
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

func CheckIps() {
	// 只检测中国、香港、新加坡IP
	lastLine := GetLogNumber("ip_line")
	ipFile, err := os.Open(ExecPath + "data/ip.merge.txt")
	if err != nil {
		log.Println("ip 文件不存在")
		return
	}
	defer ipFile.Close()

	var line string
	cursor := int64(0)
	reader := bufio.NewReader(ipFile)
	for {
		line, err = reader.ReadString('\n')
		cursor++
		if lastLine >= cursor {
			// 读取下一行
			continue
		}

		if strings.Contains(line, "中国") ||
			strings.Contains(line, "香港") ||
			strings.Contains(line, "新加坡") {
			_ = StoreLogNumber("ip_line", cursor)
			line = strings.TrimSpace(line)
			verifyIPs(line)
		}

		if err != nil {
			// 没了
			break
		}
	}
	time.Sleep(10 * time.Second)
	log.Println("finished")
}

var ipChs = make(chan string, 500)

func verifyIPs(line string) {
	ips := strings.Split(line, "|")
	if len(ips) < 2 {
		return
	}
	start, err1 := Ipv4ToLong(ips[0])
	end, err2 := Ipv4ToLong(ips[1])
	if err1 != nil || err2 != nil {
		return
	}
	for i := start; i <= end; i++ {
		ip, _ := LongToIpv4(i)
		ipChs <- ip
		go pingIP(ip)
	}
}

var counter int

func pingIP(ip string) {
	defer func() {
		<-ipChs
	}()
	// 这是一个真实IP
	counter++
	fmt.Println(counter)
	open := Ping(ip, "80")
	if open {
		DebugLog("valid_ip", ip)
		fmt.Println(ip, " 开放80")
	}
}

// get all c segments
func GetAllCSegments() {
	segments := map[string]bool{}
	ipFile, err := os.Open(ExecPath + "data/ip.merge.txt")
	if err != nil {
		log.Println("ip 文件不存在")
		return
	}
	defer ipFile.Close()

	var line string
	cursor := int64(0)
	reader := bufio.NewReader(ipFile)
	for {
		line, err = reader.ReadString('\n')
		cursor++
		if strings.Contains(line, "中国") ||
			strings.Contains(line, "香港") ||
			strings.Contains(line, "新加坡") {

			ips := strings.Split(line, "|")
			if len(ips) < 2 {
				continue
			}
			start, err1 := Ipv4ToLong(ips[0])
			end, err2 := Ipv4ToLong(ips[1])
			if err1 != nil || err2 != nil {
				continue
			}
			for i := start; i <= end; i++ {
				ip, _ := LongToIpv4(i)
				dot := strings.Split(ip, ".")
				dot[3] = "0"
				ip = strings.Join(dot, ".")
				if _, ok := segments[ip]; !ok {
					segments[ip] = true
					DebugLog("csegments.txt", ip)
				}
			}
		}

		if err != nil {
			// 没了
			break
		}
	}
	log.Println("finished")
}

func Ipv4ToLong(ip string) (uint, error) {
	p := net.ParseIP(ip).To4()
	if p == nil {
		return 0, errors.New("invalid ipv4 format")
	}

	return uint(p[0])<<24 | uint(p[1])<<16 | uint(p[2])<<8 | uint(p[3]), nil
}

func LongToIpv4(i uint) (string, error) {
	if i > math.MaxUint32 {
		return "", errors.New("beyond the scope of ipv4")
	}

	ip := make(net.IP, net.IPv4len)
	ip[0] = byte(i >> 24)
	ip[1] = byte(i >> 16)
	ip[2] = byte(i >> 8)
	ip[3] = byte(i)

	return ip.String(), nil
}
