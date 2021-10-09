package cobweb

import (
	"context"
	"crypto/tls"
	"github.com/axgle/mahonia"
	"github.com/parnurzeal/gorequest"
	"golang.org/x/net/html/charset"
	"golang.org/x/net/publicsuffix"
	"net"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"sync"
	"time"
)

type RequestData struct {
	Header     http.Header
	Request    *http.Request
	Body       string
	Status     string
	StatusCode int
	Domain     string
	Scheme     string
	IP         string
	Server     string
	ProxyIP    string
}

type Options struct {
	Timeout     time.Duration
	Method      string
	Type        string
	Query       interface{}
	Data        interface{}
	Header      map[string]string
	Proxy       bool
	ProxyIP     string
	Cookies     []*http.Cookie
	UserAgent   string
	IsMobile    bool
	Debug       bool
	DialContext func(ctx context.Context, network, addr string) (net.Conn, error)
}

var httpClient *http.Client

var mu = sync.Mutex{}

func getHttpClient() *http.Client {
	mu.Lock()
	if httpClient != nil {
		mu.Unlock()
		return httpClient
	}
	cookiejarOptions := cookiejar.Options{
		PublicSuffixList: publicsuffix.List,
	}
	jar, _ := cookiejar.New(&cookiejarOptions)

	httpClient = &http.Client{Jar: jar}
	mu.Unlock()
	return httpClient
}

/**
 * 请求网络页面，并自动检测页面内容的编码，转换成utf-8
 */
func Request(urlPath string, options *Options) (*RequestData, error) {
	if options == nil {
		options = &Options{}
	}
	if options.Timeout == 0 {
		options.Timeout = 10
	}
	if options.Method == "" {
		options.Method = "GET"
	}
	options.Method = strings.ToUpper(options.Method)

	req := gorequest.New().SetDoNotClearSuperAgent(true).TLSClientConfig(&tls.Config{InsecureSkipVerify: true}).Timeout(options.Timeout * time.Second)
	if options.Debug {
		req.SetDebug(true)
	}
	//都使用一个client
	req.Client = getHttpClient()
	if options.Type != "" {
		req = req.Type(options.Type)
	}
	proxyIP := ""
	if options.ProxyIP != "" {
		proxyIP = options.ProxyIP
		req = req.Proxy(options.ProxyIP)
	} else if options.Proxy {
		//声明这次请求需要使用代理，也有可能代理池里没有代理ip，传入空字符串不会使用代理
		// todo
	}
	if options.Cookies != nil {
		req = req.AddCookies(options.Cookies)
	}
	if options.Query != nil {
		req = req.Query(options.Query)
	}
	if options.Data != nil {
		req = req.Send(options.Data)
	}
	if options.Header != nil {
		for i, v := range options.Header {
			req = req.Set(i, v)
		}
	}

	if options.UserAgent == "" {
		options.UserAgent = getRandomAgent(options.IsMobile)
	}
	req = req.Set("User-Agent", options.UserAgent)

	if options.DialContext != nil {
		req.Transport.DialContext = options.DialContext
	}

	if options.Method == "POST" {
		req = req.Post(urlPath)
	} else {
		req = req.Get(urlPath)
	}

	resp, body, errs := req.End()
	if len(errs) > 0 {
		////如果是https,则尝试退回http请求
		//if strings.HasPrefix(urlPath, "https://") {
		//	urlPath = strings.Replace(urlPath, "https://", "http://", 1)
		//	return Request(urlPath, options)
		//}
		return &RequestData{ProxyIP: proxyIP}, errs[0]
	}
	//
	//domain, _ := url.Parse(urlPath)
	//log.Println(req.Client.Jar.Cookies(domain))

	//contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	//body = toUtf8(body, contentType)

	requestData := RequestData{
		Header:     resp.Header,
		Request:    resp.Request,
		Body:       body,
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Domain:     resp.Request.Host,
		Scheme:     resp.Request.URL.Scheme,
		Server:     resp.Header.Get("Server"),
		ProxyIP:    proxyIP,
	}

	return &requestData, nil
}

/**
 * 对外公开的编码转换接口，传入的字符串会自动检测编码，并转换成utf-8
 */
func ToUtf8(content string) string {
	return toUtf8(content, "")
}

/**
 * 内部编码判断和转换，会自动判断传入的字符串编码，并将它转换成utf-8
 * windows-1252 并不是一个具体的编码，直接拿它来转码会失败
 */
func toUtf8(content string, contentType string) string {
	var htmlEncode string
	var htmlEncode2 string
	var htmlEncode3 string
	if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
		htmlEncode = "gb18030"
	} else if strings.Contains(contentType, "big5") {
		htmlEncode = "big5"
	} else if strings.Contains(contentType, "utf-8") {
		//实际上，这里获取的编码未必是正确的，在下面还要做比对
		htmlEncode = "utf-8"
	}

	reg := regexp.MustCompile(`(?is)<meta[^>]*charset\s*=["']?\s*([A-Za-z0-9\-]+)`)
	match := reg.FindStringSubmatch(content)
	if len(match) > 1 {
		contentType = strings.ToLower(match[1])
		if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
			htmlEncode2 = "gb18030"
		} else if strings.Contains(contentType, "big5") {
			htmlEncode2 = "big5"
		} else if strings.Contains(contentType, "utf-8") {
			htmlEncode2 = "utf-8"
		}
	}

	reg = regexp.MustCompile(`(?is)<title[^>]*>(.*?)<\/title>`)
	match = reg.FindStringSubmatch(content)
	if len(match) > 1 {
		aa := match[1]
		_, contentType, _ = charset.DetermineEncoding([]byte(aa), "")
		contentType = strings.ToLower(contentType)
		if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
			htmlEncode3 = "gb18030"
		} else if strings.Contains(contentType, "big5") {
			htmlEncode3 = "big5"
		} else if strings.Contains(contentType, "utf-8") {
			htmlEncode3 = "utf-8"
		}
	}

	//fmt.Println(fmt.Sprintf("contentType:%s, htmlEncode:%s, htmlEncode2:%s, htmlEncode3:%s", contentType, htmlEncode, htmlEncode2, htmlEncode3))
	if htmlEncode3 != "" && htmlEncode2 != htmlEncode3 {
		htmlEncode2 = htmlEncode3
	}
	if htmlEncode2 != "" && htmlEncode != htmlEncode2 {
		htmlEncode = htmlEncode2
	}

	//fmt.Println(fmt.Sprintf("contentType:%s, htmlEncode:%s, htmlEncode2:%s, htmlEncode3:%s", contentType, htmlEncode, htmlEncode2, htmlEncode3))
	if htmlEncode != "" && htmlEncode != "utf-8" {
		content = Convert(content, htmlEncode, "utf-8")
	}

	return content
}

/**
 * 编码转换
 * 需要传入原始编码和输出编码，如果原始编码传入出错，则转换出来的文本会乱码
 */
func Convert(src string, srcCode string, tagCode string) string {
	if srcCode == tagCode {
		return src
	}
	srcCoder := mahonia.NewEncoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}
