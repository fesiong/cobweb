# 全网网址采集器(cobweb)
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

## 版权声明
© Fesion，tpyzlxy@163.com

Released under the [MIT License](https://github.com/fesiong/cobweb/blob/master/License)