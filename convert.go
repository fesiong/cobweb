package cobweb

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"github.com/parnurzeal/gorequest"
	"github.com/tdewolff/parse/v2/buffer"
	"golang.org/x/net/html/charset"
	"golang.org/x/net/publicsuffix"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"regexp"
	"strings"
	"sync"
	"time"
)

var DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.164 Safari/537.36"

// charsetPatternInDOMStr meta[http-equiv]元素, content属性中charset截取的正则模式.
// 如<meta http-equiv="content-type" content="text/html; charset=utf-8">
var charsetPatternInDOMStr = `charset\s*=\s*(\S*)\s*;?`

// charsetPattern 普通的MatchString可直接接受模式字符串, 无需Compile,
// 但是只能作为判断是否匹配, 无法从中获取其他信息.
var charsetPattern = regexp.MustCompile(charsetPatternInDOMStr)

// CharsetMap 字符集映射
var CharsetMap = map[string]encoding.Encoding{
	"utf-8":   unicode.UTF8,
	"gbk":     simplifiedchinese.GBK,
	"gb2312":  simplifiedchinese.GB18030,
	"gb18030": simplifiedchinese.GB18030,
	"big5":    traditionalchinese.Big5,
}

// HTMLCharacterEntitiesMap HTML 字符实体
var HTMLCharacterEntitiesMap = map[string]string{
	"\u00a0": "&nbsp;",
	"©":      "&copy;",
	"®":      "&reg;",
	"™":      "&trade;",
	"￠":      "&cent;",
	"£":      "&pound;",
	"¥":      "&yen;",
	"€":      "&euro;",
	"§":      "&sect;",
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
	DialContext func(ctx context.Context, network, addr string) (net.Conn, error)
	File        interface{}
	FileName    string
	Debug       bool
	NotFollow   bool
}

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

// Request
// 请求网络页面，并自动检测页面内容的编码，转换成utf-8
func Request(urlPath string, options *Options) (*RequestData, error) {
	if options == nil {
		options = &Options{
			Method:  "GET",
			Timeout: 10,
		}
	}
	if options.Timeout == 0 {
		options.Timeout = 10
	}
	if options.Method == "" {
		options.Method = "GET"
	}
	options.Method = strings.ToUpper(options.Method)

	req := gorequest.New().SetDebug(options.Debug).SetDoNotClearSuperAgent(true).TLSClientConfig(&tls.
		Config{InsecureSkipVerify: true}).Timeout(options.Timeout * time.Second)

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

	if options.File != nil {
		req = req.SendFile(options.File, options.FileName)
	}

	if options.Header != nil {
		for i, v := range options.Header {
			req = req.Set(i, v)
		}
	}

	req = req.Set("User-Agent", DefaultUserAgent)

	if options.DialContext != nil {
		req.Transport.DialContext = options.DialContext
	}

	if options.NotFollow {
		req = req.RedirectPolicy(func(req gorequest.Request, via []gorequest.Request) error {
			return http.ErrUseLastResponse
		})
	}

	if options.Method == "POST" {
		req = req.Post(urlPath)
	} else if options.Method == "PUT" {
		req = req.Put(urlPath)
	} else if options.Method == "DELETE" {
		req = req.Delete(urlPath)
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

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if strings.Contains(contentType, "html") || strings.Contains(contentType, "javascript") {
		// 编码处理
		charsetName, err := GetPageCharset(body, contentType)
		if err != nil {
			log.Println("获取页面编码失败: ", err.Error())
		}
		charsetName = strings.ToLower(charsetName)
		//log.Println("当前页面编码:", charsetName)
		charSet, exist := CharsetMap[charsetName]
		if !exist {
			log.Println("未找到匹配的编码")
		}
		if charSet != nil {
			utf8Coutent, err := DecodeToUTF8([]byte(body), charSet)
			if err != nil {
				log.Println("页面解码失败: ", err.Error())
			} else {
				body = string(utf8Coutent)
			}
		}
	}

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

var httpClient *http.Client

func getHttpClient() *http.Client {
	mu := sync.Mutex{}
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

// DecodeToUTF8 从输入的byte数组中按照指定的字符集解析出对应的utf8格式的内容并返回.
func DecodeToUTF8(input []byte, charset encoding.Encoding) (output []byte, err error) {
	if charset == nil || charset == unicode.UTF8 {
		output = input
		return
	}
	reader := transform.NewReader(bytes.NewReader(input), charset.NewDecoder())
	output, err = io.ReadAll(reader)
	if err != nil {
		return
	}
	return
}

// EncodeFromUTF8 将输入的utf-8格式的byte数组中按照指定的字符集编码并返回
func EncodeFromUTF8(input []byte, charset encoding.Encoding) (output []byte, err error) {
	if charset == nil || charset == unicode.UTF8 {
		output = input
		return
	}
	reader := transform.NewReader(bytes.NewReader(input), encoding.ReplaceUnsupported(charset.NewEncoder()))
	output, err = io.ReadAll(reader)
	if err != nil {
		return
	}
	return
}

// GetPageCharset 解析页面, 从中获取页面编码信息
func GetPageCharset(content, contentType string) (charSet string, err error) {
	//log.Println("服务器返回编码：", contentType)
	if contentType != "" {
		matchedArray := charsetPattern.FindStringSubmatch(strings.ToLower(contentType))
		if len(matchedArray) > 1 {
			for _, matchedItem := range matchedArray[1:] {
				if strings.ToLower(matchedItem) != "utf-8" {
					charSet = matchedItem
					return
				}
			}
		}
	}
	//log.Println("继续查找编码1")
	var checkType string
	reg := regexp.MustCompile(`(?is)<title[^>]*>(.*?)<\/title>`)
	match := reg.FindStringSubmatch(content)
	if len(match) > 1 {
		_, checkType, _ = charset.DetermineEncoding([]byte(match[1]), "")
		//log.Println("Title解析编码：", checkType)
		if checkType == "utf-8" {
			charSet = checkType
			return
		}
	}
	//log.Println("继续查找编码2")
	reg = regexp.MustCompile(`(?is)<meta[^>]*charset\s*=["']?\s*([\w\d\-]+)`)
	match = reg.FindStringSubmatch(content)
	if len(match) > 1 {
		charSet = match[1]
		return
	}
	//log.Println("找不到编码")
	charSet = "utf-8"
	return
}

func GetURLData(url, refer string, timeout int) (*http.Response, []byte, error) {
	client := &http.Client{}
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	} else if timeout < 0 {
		client.Timeout = 10 * time.Second
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", DefaultUserAgent)
	req.Header.Set("Referer", refer)

	resp, err := client.Do(req)
	if err != nil {
		// 如果第一次失效，则重新尝试一次
		if timeout != -1 {
			return GetURLData(url, refer, -1)
		}
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "html") || strings.Contains(contentType, "javascript") {
		// 编码处理
		charsetName, err := GetPageCharset(string(body), contentType)
		if err != nil {
			log.Println("获取页面编码失败: ", err.Error())
		}
		charsetName = strings.ToLower(charsetName)
		charSet, exist := CharsetMap[charsetName]
		if !exist {
			log.Println("未找到匹配的编码")
		}
		if charSet != nil {
			utf8Coutent, err := DecodeToUTF8(body, charSet)
			if err != nil {
				log.Println("页面解码失败: ", err.Error())
			} else {
				body = utf8Coutent
			}
		}
	}

	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return resp, body, nil
}

func PostURLData(url string, data interface{}, timeout int) (*http.Response, []byte, error) {
	client := &http.Client{}
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	} else if timeout < 0 {
		client.Timeout = 10 * time.Second
	}
	buf, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, buffer.NewReader(buf))
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	resp.Body = io.NopCloser(bytes.NewBuffer(body))

	return resp, body, nil
}

// ReplaceHTMLCharacterEntities 替换页面中html实体字符, 以免写入文件时遇到不支持的字符
func ReplaceHTMLCharacterEntities(input string, charset encoding.Encoding) (output string) {
	if charset == nil || charset == unicode.UTF8 {
		output = input
		return
	}
	output = html.UnescapeString(input)
	for char, entity := range HTMLCharacterEntitiesMap {
		output = strings.Replace(output, char, entity, -1)
	}
	return
}
