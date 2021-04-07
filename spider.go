package cobweb

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/fesiong/goproject/convert"
	"log"
	"net"
	"regexp"
	"strings"
	"sync"
	"time"
)

var MaxChan = 50
var waitGroup sync.WaitGroup
var ch = make(chan string, MaxChan)

/**
 * 循环获取数据
 */
func StartSpider() {
	//for i := 0; i < MaxChan; i++ {
	//	ch <- i
	//	waitGroup.Add(1)
	//	go SingleSpider()
	//}

	SingleSpider()
	waitGroup.Wait()
	log.Println("执行结束")
}

func SingleSpider() {
	var websites []Website
	var counter int64
	DB.Model(&Website{}).Where("`status` = 0").Limit(MaxChan * 10).Count(&counter).Find(&websites)
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
func SingleData() {
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
		DB.Model(&website).UpdateColumn("status", 2)
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
				log.Println(fmt.Sprintf("入库：%s", v.Domain))
				DB.Exec("insert into website(`domain`, `scheme`,`title`) select ?,?,? from dual where not exists(select id from website where `domain` = ?)", v.Domain, v.Scheme, v.Title, v.Domain)
			}
		}
	}

	SingleData()
}

func SingleData2(website Website) {
	defer func() {
		waitGroup.Done()
		<-ch
	}()
	//锁定当前数据
	DB.Model(&website).Where("`domain` = ?", website.Domain).UpdateColumn("status", 2)
	log.Println(fmt.Sprintf("开始采集：%s://%s", website.Scheme, website.Domain))
	err := website.GetWebsite()
	if err == nil {
		website.Status = 1
	} else {
		website.Status = 3
	}
	log.Println(fmt.Sprintf("入库2：%s", website.Domain))
	DB.Where("`domain` = ?", website.Domain).Save(&website)
	if len(website.Links) > 0 {
		for _, v := range website.Links {
			webData := Website{
				Url:    v.Url,
				Domain: v.Domain,
				Scheme: v.Scheme,
				Title:  v.Title,
			}
			DB.Where("`domain` = ?", webData.Domain).FirstOrCreate(&webData)
			log.Println(fmt.Sprintf("入库：%s", v.Domain))
			//DB.Exec("insert into website(`domain`, `scheme`,`title`) select ?,?,? from dual where not exists(select domain from website where `domain` = ?)", v.Domain, v.Scheme, v.Title, v.Domain)
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
	requestData, err := convert.Request(website.Url)
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
				hosts := strings.Split(host, ".")
				if len(hosts) < 2 || len(hosts) > 3 && hosts[0] != "www" {
					//refuse
				} else {
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
	}

	return links
}

func init() {
	initDB()

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
