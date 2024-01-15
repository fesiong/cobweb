package cobweb

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gorm.io/gorm/clause"
	"log"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var MaxChan = 20
var ch = make(chan string, MaxChan)

var existsDomain = &sync.Map{}
var runMap = &sync.Map{}
var topDomains = &sync.Map{}

/*
StartSpider
  - 循环获取数据
*/
func StartSpider() {
	var websites []*Website
	lastId := 0
	for {
		DB.Where("id > ?", lastId).Order("id asc").Limit(50000).Find(&websites)
		if len(websites) == 0 {
			break
		}
		lastId = websites[len(websites)-1].ID
		log.Println("load: ", lastId)
		for _, v := range websites {
			item, ok := topDomains.Load(v.TopDomain)
			if ok {
				topDomains.Store(v.TopDomain, item.(int)+1)
			} else {
				topDomains.Store(v.TopDomain, 1)
			}

			existsDomain.Store(v.Domain, true)
			if v.Status == 0 {
				runMap.Store(v.Domain, v.Scheme)
			}
		}
	}

	for {
		SingleSpider()
		log.Println("等待数据中，10秒后重试")
		time.Sleep(2 * time.Second)
	}
	log.Println("执行结束")
}

var lastGetId int64

func SingleSpider() {
	lastGetId = GetLogNumber("last-get-id")
	var waitGroup sync.WaitGroup
	runMap.Range(func(key interface{}, value interface{}) bool {
		domain := key.(string)
		scheme := value.(string)

		v := Website{
			Domain: domain,
			Scheme: scheme,
		}

		ch <- v.Domain
		waitGroup.Add(1)
		go func(vv Website) {
			defer func() {
				waitGroup.Done()
				<-ch
			}()
			SingleData2(vv)
		}(v)

		return true
	})

	var websites []Website
	DB.Model(&Website{}).Where("id > ? and `status` = 0", lastGetId).Limit(MaxChan * 10).Order("id asc").Find(&websites)
	if len(websites) > 0 {
		lastGetId = int64(websites[len(websites)-1].ID)
		for _, v := range websites {
			ch <- v.Domain
			waitGroup.Add(1)
			go func(vv Website) {
				defer func() {
					waitGroup.Done()
					<-ch
				}()
				SingleData2(vv)
			}(v)
		}
	} else {
		//读取完了，重头开始吗？
		lastGetId = 0
		DB.Model(&Website{}).Where("`id` > 0").UpdateColumn("status", 0)
	}

	StoreLogNumber("last-get-id", lastGetId)
	waitGroup.Wait()
}

func SingleData2(website Website) {
	//锁定当前数据
	DB.Model(&Website{}).Where("`domain` = ?", website.Domain).UpdateColumn("status", 2)
	log.Println(fmt.Sprintf("开始采集：%s://%s", website.Scheme, website.Domain))
	err := website.GetWebsite()
	if err == nil {
		website.Status = 1
	} else {
		website.Status = 3
	}
	log.Println(fmt.Sprintf("入库2：%s", website.Domain))
	DB.Where("`domain` = ?", website.Domain).Updates(&website)
	// 同时写入data
	contentData := WebsiteData{
		ID:      website.ID,
		Content: website.Content,
	}
	DB.Where("`id` = ?", contentData.ID).FirstOrCreate(&contentData)
	// end
	if len(website.Links) > 0 {
		for _, v := range website.Links {
			//如果超过了5个子域名，则直接抛弃
			item, itemOk := topDomains.Load(v.TopDomain)
			if itemOk {
				if item.(int) >= 4 {
					//跳过这个记录
					continue
				}
			}
			if _, ok := existsDomain.Load(v.Domain); ok {
				continue
			}
			if itemOk {
				topDomains.Store(v.TopDomain, item.(int)+1)
			} else {
				topDomains.Store(v.TopDomain, 1)
			}
			existsDomain.Store(v.Domain, true)
			runMap.Store(v.Domain, v.Scheme)
			webData := Website{
				Url:       v.Url,
				Domain:    v.Domain,
				TopDomain: v.TopDomain,
				Scheme:    v.Scheme,
				Title:     v.Title,
			}
			DB.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Where("`domain` = ?", webData.Domain).Create(&webData)
			log.Println(fmt.Sprintf("入库：%s", v.Domain))
		}
	}
}

/*
GetWebsite
  - 一个域名数据抓取
*/
func (website *Website) GetWebsite() error {
	if website.Url == "" {
		website.Url = website.Scheme + "://" + website.Domain
	}
	ops := &Options{}
	if ProxyValid {
		ops.ProxyIP = JsonData.DBConfig.Proxy
	}
	requestData, err := Request(website.Url, ops)
	if err != nil {
		log.Println(err)
		return err
	}
	// 注入内容
	website.Content = requestData.Body

	if requestData.Domain != "" {
		website.Domain = requestData.Domain
		website.TopDomain = getTopDomain(website.Domain)
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

	//尝试判断cms
	website.Cms = getCms(requestData.Body)

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
	if len(match) > 1 { //
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
				if len(hosts) < 2 || (len(hosts) > 3 && hosts[0] != "www") || (len(hosts) == 3 && hosts[0] != "www" && len(hosts[1]) > 4) {
					//refuse
				} else {
					if !existsLinks[host] {
						//去重
						existsLinks[host] = true
						links = append(links, Link{
							Title:     title,
							Url:       scheme + "://" + host,
							Domain:    host,
							TopDomain: getTopDomain(host),
							Scheme:    scheme,
						})
					}
				}
			}
		}
	}

	return links
}

func getTopDomain(domain string) string {
	//第二后缀部分，com,org,gov,net,和cn所有2个字母的部分，可以认为是双后缀
	items := strings.Split(domain, ".")
	if len(items) <= 2 {
		//只有2个，就是顶级了
		return domain
	}
	//先截取成三个
	items = items[len(items)-3:]
	//先判断是否是双后缀，
	if items[1] == "com" || items[1] == "org" || items[1] == "gov" || items[1] == "net" || (len(items[1]) == 2 && items[2] == "cn") {
		//认为是双后缀
		return strings.Join(items, ".")
	}
	//一般的域名
	items = items[1:]
	return strings.Join(items, ".")
}

func getCms(content string) string {
	//discuz
	if strings.Contains(content, "Discuz!") || strings.Contains(content, "forum.php") {
		return "discuz"
	}
	if strings.Contains(content, "/a/") || strings.Contains(content, "/plus/") {
		return "dedecms"
	}
	if strings.Contains(content, "/metinfo/") {
		return "mitInfo"
	}
	if strings.Contains(content, "/e/") {
		return "empirecms"
	}
	if strings.Contains(content, "/wp-/") {
		return "wordpress"
	}
	if strings.Contains(content, "/templates/default/") && strings.Contains(content, "/index.php?ac=news") {
		return "wodecms"
	}
	if strings.Contains(content, "/html/") || strings.Contains(content, "/template") || strings.Contains(content, "/uploadfiles/") {
		return "cms"
	}

	return ""
}

func init() {
	initDB()

	websites := []Website{
		{Scheme: "https", Domain: "www.hao123.com", TopDomain: "hao123.com"},
		{Scheme: "https", Domain: "www.2345.com", TopDomain: "2345.com"},
		{Scheme: "https", Domain: "hao.360.com", TopDomain: "360.com"},
	}

	for _, website := range websites {
		DB.Where("`domain` = ?", website.Domain).FirstOrCreate(&website)
	}

	buf, err := os.ReadFile(fmt.Sprintf("%surls.txt", ExecPath))
	if err == nil {
		links := strings.Split(string(buf), "\n")
		for _, v := range links {
			parsed, err := url.Parse(strings.TrimSpace(v))
			if err == nil {
				website := &Website{
					Scheme:    parsed.Scheme,
					Domain:    parsed.Hostname(),
					TopDomain: getTopDomain(parsed.Hostname()),
				}
				DB.Where("`domain` = ?", website.Domain).FirstOrCreate(&website)
			}
		}
	}

	//释放所有status=2的数据
	DB.Model(&Website{}).Where("status = 2").UpdateColumn("status", 0)
}
