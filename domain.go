package cobweb

import (
	"bufio"
	"fmt"
	"gorm.io/gorm/clause"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var otherCount = 0

func CheckDomains() {
	// 只检测中国、香港、新加坡IP
	lastLine := GetLogNumber("csegments")
	ipFile, err := os.Open(ExecPath + "data/csegments.txt")
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
		line = strings.TrimSpace(line)

		getDomains(line)

		_ = StoreLogNumber("csegments", cursor)

		if err != nil {
			// 没了
			break
		}
	}
	time.Sleep(10 * time.Second)
	log.Println("finished")
}

func getDomains(ip24 string) {
	fromUrl := fmt.Sprintf("https://chapangzhan.com/%s/24", ip24)
	log.Println(fromUrl)
	ops := &Options{}
	if ProxyValid {
		//	ops.ProxyIP = JsonData.DBConfig.Proxy
	}
	requestData, err := Request(fromUrl, ops)
	if err != nil {
		log.Println("错误，退出", err)
		os.Exit(1)
	}
	if requestData.StatusCode != 200 {
		if otherCount >= 5 {
			log.Println("错误，退出", requestData.StatusCode)
			log.Println(requestData.Body)
			os.Exit(1)
		}
		otherCount++
		return
	}
	//解析文档内容
	re, _ := regexp.Compile(`<a href="(https://ipchaxun\.com/([0-9.]+)/)"`)
	matches := re.FindAllStringSubmatch(requestData.Body, -1)
	for _, match := range matches {
		getDomainByIp(match[1], fromUrl)
	}
}

func getDomainByIp(ipLink, refer string) {
	log.Println(ipLink)
	ops := &Options{
		Header: map[string]string{
			"Referer": refer,
		},
	}
	if ProxyValid {
		//	ops.ProxyIP = JsonData.DBConfig.Proxy
	}
	requestData, err := Request(ipLink, ops)
	if err != nil {
		log.Println("错误，退出", err, ipLink)
		os.Exit(1)
	}
	if requestData.StatusCode != 200 {
		if otherCount >= 5 {
			log.Println("错误，退出", requestData.StatusCode)
			log.Println(requestData.Body)
			os.Exit(1)
		}
		otherCount++
		return
	}
	re, _ := regexp.Compile(`<a href="/([a-z0-9\-.]+\.[a-z]+)/"`)

	matches := re.FindAllStringSubmatch(requestData.Body, -1)

	for _, match := range matches {
		// 不插入IP
		webData := Website{
			Domain:    match[1],
			TopDomain: getTopDomain(match[1]),
			Scheme:    "http",
			Title:     "",
		}
		rowAffected := DB.Clauses(clause.OnConflict{
			DoNothing: true,
		}).Where("`domain` = ?", match[1]).Create(&webData).RowsAffected
		if rowAffected > 0 {
			log.Println(fmt.Sprintf("入库：%s", match[1]))
		}
	}
}

func GetContents() {
	lastId := GetLogNumber("last-content-id")
	wg := sync.WaitGroup{}
	for {
		var websites []Website
		DB.Model(&Website{}).Where("id > ? and `status` = 0", lastId).Limit(MaxChan * 10).Order("id asc").Find(&websites)
		if len(websites) == 0 {
			break
		}
		lastId = int64(websites[len(websites)-1].ID)
		for _, v := range websites {
			ch <- v.Domain
			wg.Add(1)
			go func(vv Website) {
				defer func() {
					wg.Done()
					<-ch
				}()
				getAndStoreContent(vv)
			}(v)
		}
		StoreLogNumber("last-content-id", lastId)
	}
	wg.Wait()
	log.Println("finished")
}

func getAndStoreContent(website Website) error {
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

	contentData := WebsiteData{
		ID:      website.ID,
		Content: website.Content,
	}
	DB.Where("`id` = ?", contentData.ID).FirstOrCreate(&contentData)

	return nil
}
