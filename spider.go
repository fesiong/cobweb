package cobweb

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/parnurzeal/gorequest"
	"golang.org/x/net/html/charset"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

var MaxChan = 100
var waitGroup sync.WaitGroup
var ch = make(chan string, MaxChan)

/**
 * 循环获取数据
 */
func StartSpider(){
	//for i := 0; i < MaxChan; i++ {
	//	ch <- i
	//	waitGroup.Add(1)
	//	go SingleSpider()
	//}

	SingleSpider()
	waitGroup.Wait()
	log.Println("执行结束")
}

func SingleSpider(){
	var websites []Website
	var counter int
	DB.Model(&Website{}).Where("`status` = 0").Limit(MaxChan*10).Count(&counter).Find(&websites)
	if counter > 0 {
		for _, v := range websites {
			ch <- v.Domain
			waitGroup.Add(1)
			go SingleData2(v)
		}
	} else {
		log.Println("等待数据中，10秒后重试")
		time.Sleep(10 * time.Second)
	}
	SingleSpider()
}

/**
 * 单个执行
 */
func SingleData(){
	counter := 0
	var website Website
	err := DB.Where("`status` = 0").First(&website).Error
	if err != nil {
		counter++
		if counter > 10 {
			return
		}

		log.Println("等待数据中，10秒后重试")
		time.Sleep(10 * time.Second)
	} else {
		//锁定当前数据
		DB.Model(&website).Where("`id` = ?", website.ID).Update("status", 2)
		log.Println(fmt.Sprintf("开始采集：%s://%s", website.Scheme, website.Domain))
		err = website.GetWebsite()
		if err == nil {
			website.Status = 1
		} else {
			website.Status = 3
		}
		DB.Save(&website)
		if len(website.Links) > 0 {
			for _, v := range website.Links {
				//webData := Website{
				//	Url:    v.Url,
				//	Domain: v.Domain,
				//	Scheme: v.Scheme,
				//	Title:  v.Title,
				//}
				log.Println(fmt.Sprintf("入库：%d：%s",website.ID, v.Domain))
				DB.Exec("insert into website(`domain`, `scheme`,`title`) select ?,?,? from dual where not exists(select id from website where `domain` = ?)", v.Domain, v.Scheme, v.Title, v.Domain)
			}
		}
	}

	SingleData()
}

func SingleData2(website Website){
	defer func() {
		waitGroup.Done()
		<-ch
	}()
	//锁定当前数据
	DB.Model(&website).Where("`id` = ?", website.ID).Update("status", 2)
	log.Println(fmt.Sprintf("开始采集：%s://%s", website.Scheme, website.Domain))
	err := website.GetWebsite()
	if err == nil {
		website.Status = 1
	} else {
		website.Status = 3
	}
	log.Println(fmt.Sprintf("入库2：%d：%s",website.ID, website.Domain))
	DB.Save(&website)
	if len(website.Links) > 0 {
		for _, v := range website.Links {
			//webData := Website{
			//	Url:    v.Url,
			//	Domain: v.Domain,
			//	Scheme: v.Scheme,
			//	Title:  v.Title,
			//}
			log.Println(fmt.Sprintf("入库：%d：%s",website.ID, v.Domain))
			DB.Exec("insert into website(`domain`, `scheme`,`title`) select ?,?,? from dual where not exists(select id from website where `domain` = ?)", v.Domain, v.Scheme, v.Title, v.Domain)
		}
	}
}

/**
 * 一个域名数据抓取
 */
func (website *Website) GetWebsite() error {
	if website.Url == "" {
		website.Url = website.Scheme + "://" + website.Domain
	}
	requestData, err := Request(website.Url)
	if err != nil {
		log.Println(err)
		return err
	}
	if requestData.Domain != "" {
		website.Domain = requestData.Domain
	}
	if requestData.Scheme != "" {
		website.Scheme = requestData.Scheme
	}
	website.Server = requestData.Server

	//获取IP
	conn, err := net.ResolveIPAddr("ip", website.Domain)
	if err == nil {
		website.IP = conn.String()
	}

	//先删除一些不必要的标签
	re, _ := regexp.Compile("\\<style[\\S\\s]+?\\</style\\>")
	requestData.Body = re.ReplaceAllString(requestData.Body, "")
	re, _ = regexp.Compile("\\<script[\\S\\s]+?\\</script\\>")
	requestData.Body = re.ReplaceAllString(requestData.Body, "")
	//解析文档内容
	htmlR := strings.NewReader(requestData.Body)
	doc, err := goquery.NewDocumentFromReader(htmlR)
	if err != nil {
		fmt.Println(err)
		return err
	}
	contentText := doc.Text()
	contentText = strings.ReplaceAll(contentText, "\n", " ")
	contentText = strings.ReplaceAll(contentText, "\r", " ")
	contentText = strings.ReplaceAll(contentText, "\t", " ")
	website.Title = doc.Find("title").Text()
	desc, exists := doc.Find("meta[name=description]").Attr("content")
	if exists {
		website.Description = desc
	} else {
		website.Description = strings.ReplaceAll(contentText, " ", "")
	}
	nameRune := []rune(website.Description)
	curLen := len(nameRune)
	if curLen > 200 {
		website.Description = string(nameRune[:200])
	}
	nameRune = []rune(website.Title)
	curLen = len(nameRune)
	if curLen > 200 {
		website.Title = string(nameRune[:200])
	}
	//尝试获取微信
	reg := regexp.MustCompile(`(?i)(微信|微信客服|微信号|微信咨询|微信服务)\s*(:|：|\s)\s*([a-z0-9\-_]{4,30})`)
	match := reg.FindStringSubmatch(contentText)
	if len(match) > 1 {
		website.WeChat = match[3]
	}
	//尝试获取QQ
	reg = regexp.MustCompile(`(?i)(QQ|QQ客服|QQ号|QQ号码|QQ咨询|QQ联系|QQ交谈)\s*(:|：|\s)\s*([0-9]{5,12})`)
	match = reg.FindStringSubmatch(contentText)
	if len(match) > 1 {
		website.QQ = match[3]
	}
	//尝试获取电话
	reg = regexp.MustCompile(`([0148][1-9][0-9][0-9\-]{4,15})`)
	match = reg.FindStringSubmatch(contentText)
	if len(match) > 1 {
		website.Cellphone = match[1]
	}
	website.Links = CollectLinks(doc)

	return nil
}

/**
 * 读取页面链接
 */
func CollectLinks(doc *goquery.Document) []Link {
	var links []Link
	aLinks := doc.Find("a")
	//读取所有连接
	existsLinks := map[string]bool{}
	for i := range aLinks.Nodes {
		href, exists := aLinks.Eq(i).Attr("href")
		title := strings.TrimSpace(aLinks.Eq(i).Text())
		if exists {
			scheme, host := ParseDomain(href)
			if host != "" && scheme != "" {
				if !existsLinks[host] {
					//去重
					existsLinks[host] = true
					links = append(links, Link{
						Title:  title,
						Url:    scheme + "://" + host,
						Domain: host,
						Scheme: scheme,
					})
				}
			}
		}
	}

	return links
}

/**
 * 请求域名返回数据
 */
func Request(urlPath string) (*RequestData, error) {
	resp, body, errs := gorequest.New().Timeout(90 * time.Second).Get(urlPath).End()
	if len(errs) > 0 {
		//如果是https,则尝试退回http请求
		if strings.HasPrefix(urlPath, "https") {
			urlPath = strings.Replace(urlPath, "https://", "http://", 1)
			return Request(urlPath)
		}
		return nil, errs[0]
	}
	defer resp.Body.Close()
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	var htmlEncode string

	if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
		htmlEncode = "gb18030"
	} else if strings.Contains(contentType, "big5") {
		htmlEncode = "big5"
	}

	if htmlEncode == "" {
		//先尝试读取charset
		reg := regexp.MustCompile(`(?is)<meta[^>]*charset\s*=["']?\s*([A-Za-z0-9\-]+)`)
		match := reg.FindStringSubmatch(body)
		if len(match) > 1 {
			contentType = strings.ToLower(match[1])
			if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
				htmlEncode = "gb18030"
			} else if strings.Contains(contentType, "big5") {
				htmlEncode = "big5"
			}
		}
		if htmlEncode == "" {
			reg = regexp.MustCompile(`(?is)<title[^>]*>(.*?)<\/title>`)
			match = reg.FindStringSubmatch(body)
			if len(match) > 1 {
				aa := match[1]
				_, contentType, _ = charset.DetermineEncoding([]byte(aa), "")
				htmlEncode = strings.ToLower(htmlEncode)
				if strings.Contains(contentType, "gbk") || strings.Contains(contentType, "gb2312") || strings.Contains(contentType, "gb18030") || strings.Contains(contentType, "windows-1252") {
					htmlEncode = "gb18030"
				} else if strings.Contains(contentType, "big5") {
					htmlEncode = "big5"
				}
			}
		}
	}
	if htmlEncode != "" {
		body = ConvertToString(body, htmlEncode, "utf-8")
	}

	requestData := RequestData{
		Body:   body,
		Domain: resp.Request.Host,
		Scheme: resp.Request.URL.Scheme,
		Server: resp.Header.Get("Server"),
	}

	return &requestData, nil
}

func init(){
	websites := []Website{
		{Scheme: "https", Domain: "www.hao123.com"},
		{Scheme: "https", Domain: "www.2345.com"},
		{Scheme: "https", Domain: "hao.360.com"},
	}

	for _, website := range websites {
		DB.Where("`domain` = ?", website.Domain).FirstOrCreate(&website)
	}
	//释放所有status=2的数据
	DB.Model(&Website{}).Where("status = 2").Update("status", 0)
}