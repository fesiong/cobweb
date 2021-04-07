package cobweb

import (
	"fmt"
	"github.com/fesiong/goproject/convert"
	"log"
	"testing"
)

// Test for Request.
func TestRequest(t *testing.T) {
	link := "http://jh.517sichuan.com/"
	resp, err := convert.Request(link)
	if err != nil {
		t.Error(err.Error())
	}
	t.Skip(resp)
}

func TestWebsite_GetWebsite(t *testing.T) {
	fmt.Println("hello")
	//cobweb.StartSpider()
	//
	website := Website{
		Domain: "publicopinion.legaldaily.com.cn",
		Scheme: "https",
	}

	err := website.GetWebsite()
	log.Println(website, err)
}
