# 全网网址采集器(cobweb)
这是一个由golang编写的全网网址采集器，可用自动爬取可触及的所有网站信息。该网址采集器会自动采集并分析网站的标题、站点描述、微信、QQ、联系电话、网站所用的运行环境、ip信息等，甚至是网站所用的框架。  
<del>全新升级，使用sqlite作为数据库，不再需要安装mysql了，直接运行可执行文件就可以抓内容了。</del>

## 为什么会有这个全网网址采集器
* 因为我想收集现在全网的网址，并分析网站数据。

## 全网网址采集器能采集哪些内容
本采集器可以采集到的的内容有：文章标题、文章关键词、文章描述、文章详情内容、文章作者、文章发布时间、文章浏览量。

##全网网址采集器可用在哪里运行
本采集器可用运行在 Windows系统、Mac 系统、Linux系统（Centos、Ubuntu等），可用下载编译好的程序直接执行，也可以下载源码自己编译。

## 如何安装使用
* 下载可执行文件  
  请从Releases 中根据你的操作系统下载最新版的可执行文件，解压后，双击运行可执行文件即可开始采集之旅。  
  下载文件后，请编辑config.sample.json,将mysql配置改成你的mysql信息。然后重命名为config.json。即可允许程序。
<del>* 采集回来的数据存放在可执行文件目录下的cobweb.db 中。这个文件可以使用 Navicat 打开。打开的时候，选择sqlite3数据库。</del>
* 自助编译  
  先clone代码到本地，本地安装go运行环境，在cobweb目录下打开cmd/Terminal命令行窗口，执行命。如果你没配置代理的话，还需要新设置go的代理
```shell script
go env -w GOPROXY=https://goproxy.cn,direct
```
  最后执行下面命令  
```shell script
go mod tidy
go mod vendor
go build app/main.go
## 跨平台编译Windows版本
CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 go build -x -v -ldflags "-s -w" -o cobweb.exe ./app/main.go
```

## 版权声明
© Fesion，tpyzlxy@163.com

Released under the [MIT License](https://github.com/fesiong/cobweb/blob/master/License)