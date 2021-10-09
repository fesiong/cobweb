package cobweb

import (
	"fmt"
	"log"
	"strings"
	"testing"
)

// Test for Request.
func TestRequest(t *testing.T) {
	//link := "http://jh.517sichuan.com/"
	//resp, err := Request(link, nil)
	//if err != nil {
	//	t.Error(err.Error())
	//}
	//t.Skip(resp)
	//suborderAmount := int64(36630)
	//amount := int64(36340)
	//single := int64(24730)
	//aa := suborderAmount*100000 / amount * single
	//aaa := aa / 100000
	//bb := float64(suborderAmount)/float64(amount)*float64(single)
	//bbb := int64(math.Round(bb))
	//log.Println(aa, aaa, bb, bbb)

	result := strings.TrimSuffix("我是中国人人民", "人民")
	log.Println(result)
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
