package main

import (
	"fmt"
	"log"
	"cobweb"
)

func main() {
	fmt.Println("hello")
	//cobweb.StartSpider()
	//
	website := cobweb.Website{
		Domain:      "publicopinion.legaldaily.com.cn",
		Scheme:      "https",
	}

	err := website.GetWebsite()
	log.Println(website, err)
}



