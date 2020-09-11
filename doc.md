# 如何获取全网可访问的所有网站网址和网站信息呢

今天由于有一个小程序项目，是专门给织梦dedecms网站、WordPress网站做小程序制作免费小程序的。但是手上织梦网站和WordPress网站用户数量都不是很多，很好的项目却没有触及到用户，没有能给网站带来好处，于是就想，能不能收集现在网上所有的织梦网站和WordPress网站，并且获取他们的邮箱、QQ、微信、电话等有用信息呢？

带着疑问百度了一番，没有发现现成的可用数据，可是小程序项目还得往前推呢，等着用户来使用呢？既然网上没有现成的，要不就自己写一个吧。于是就有了这个cobweb全网网址采集器。

## 全网网址采集器是什么？
这是一个由golang编写的全网网址采集器，可用自动爬取可触及的所有网站信息。该网址采集器会自动采集并分析网站的标题、站点描述、微信、QQ、联系电话、网站所用的运行环境、ip信息等，甚至是网站所用的框架。


## 为什么会有这个全网网址采集器
* 因为我想收集现在全网的网址，并分析网站数据。

## 全网网址采集器能采集哪些内容
本采集器可以采集到的的内容有：文章标题、文章关键词、文章描述、文章详情内容、文章作者、文章发布时间、文章浏览量。

##全网网址采集器可用在哪里运行
本采集器可用运行在 Windows系统、Mac 系统、Linux系统（Centos、Ubuntu等），可用下载编译好的程序直接执行，也可以下载源码自己编译。

## 如何安装使用
* 下载可执行文件  
  请从Releases 中根据你的操作系统下载最新版的可执行文件，解压后，重命名config.dist.json为config.json，打开config.json，修改mysql部分的配置，填写为你的mysql地址、用户名、密码、数据库信息，新建cobweb数据库，导入mysql.sql到填写的数据库中，然后双击运行可执行文件即可开始采集之旅。
* 自助编译  
  先clone代码到本地，本地安装go运行环境，在cobweb目录下打开cmd/Terminal命令行窗口，执行命。如果你没配置代理的话，还需要新设置go的代理
```shell script
go env -w GOPROXY=https://goproxy.cn,direct
```
  最后执行下面命令  
```shell script
go mod tidy
go mod vendor
go build
```
编译结束后，配置config。重命名config.dist.json为config.json，打开config.json，修改mysql部分的配置，填写为你的mysql地址、用户名、密码、数据库信息，新建cobweb数据库，导入mysql.sql到填写的数据库中，然后双击运行可执行文件即可开始采集之旅。

### config.json配置说明
```
{
  "mysql": { //数据库配置
    "Database": "spider",
    "User": "root",
    "Password": "root",
    "Charset": "utf8mb4",
    "Host": "127.0.0.1",
    "TablePrefix": "",
    "Port": 3306,
    "MaxIdleConnections": 1000,
    "MaxOpenConnections": 100000
  }
}
```

## 全网网址采集器运行原理分析

### 多线程（多协程）同时执行
全网网址采集器利用了golang得天独厚的并行任务优势，同时开启多个协程，可以做到比常规轻易得手的php采集代码快10倍~100倍，甚至更快。当然更快的采集速度还需要依靠你本地的网速，你家开的是500M带宽的话，开1000个协程都是可以的。

相关代码部分
```go
var MaxChan = 100
var waitGroup sync.WaitGroup
var ch = make(chan string, MaxChan)

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
```

### 采集数据锁，最大限度保证数据不被多次执行
为了防止数据被多次执行，采集过程中还采用了数据锁，让下一次采集的时候，不会读到相同的数据。

相关代码
```go
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
```

### 快速入库，直接执行原生sql语句
本来代码中采用的是gorm的orm形式插入数据，结果发现使用orm的话，插入一条网址需要执行3条sql 语句，这在并行执行中，每多执行一条sql都是浪费，因此，采用了原生sql来插入数据，每次只需要执行一条数据即可。

相关源码
```go
DB.Exec("insert into website(`domain`, `scheme`,`title`) select ?,?,? from dual where not exists(select id from website where `domain` = ?)", v.Domain, v.Scheme, v.Title, v.Domain)
```

### 网站编码自动识别并转换为utf-8
由于是采集全网的网址，并且要从网站内容里面分析有用的数据，并且golang只支持utf-8，什么乱七八糟的编码格式网站内容，都需要转换为utf-8，因此编码转换这一步必不可少，全网网址采集器内置了功能强大的编码转换器，不用担心编码转换问题，同时兼容多种编码格式。

相关代码
```go
    contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	log.Println(contentType)
	var htmlEncode string

	if contentType == "" {
		//先尝试读取charset
		reg := regexp.MustCompile(`(?is)charset=["']?\s*([a-z0-9\-]+)`)
		match := reg.FindStringSubmatch(body)
		if len(match) > 1 {
			htmlEncode = strings.ToLower(match[1])
			if htmlEncode != "utf-8" && htmlEncode != "utf8" {
				body = ConvertToString(body, "gbk", "utf-8")
			}
		} else {
			reg = regexp.MustCompile(`(?is)<title[^>]*>(.*?)<\/title>`)
			match = reg.FindStringSubmatch(body)
			if len(match) > 1 {
				aa := match[1]
				_, htmlEncode, _ = charset.DetermineEncoding([]byte(aa), "")
				if htmlEncode != "utf-8" {
					body = ConvertToString(body, "gbk", "utf-8")
				}
			}
		}
	} else if !strings.Contains(contentType, "utf-8") {
		body = ConvertToString(body, "gbk", "utf-8")
	}
```

### 网站内容自动提取
网站信息抓取回来了，还是html状态，需要从中提取出有用的信息，才达到我们想要的最终目的，内置了qq采集、微信采集、电话采集等功能。

相关代码
```go
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
```

## 使用了哪些开源项目
全网网址采集器采用了两个非常有名的开源项目，一个是用于网站内容抓取的项目gorequest，另一个是用于网站内容分析的项目goquery。两个项目共同组成了采集器的核心功能。

如果你对采集器的原理有更大的兴趣，可以直接拜读存放在GitHub上的源码：[https://github.com/fesiong/cobweb](https://github.com/fesiong/cobweb) 