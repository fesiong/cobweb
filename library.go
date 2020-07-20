package cobweb

import (
	"github.com/axgle/mahonia"
	"net"
	"net/url"
	"strings"
)

func ConvertToString(src string, srcCode string, tagCode string) string {
	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func ParseDomain(urlPath string) (scheme string, domain string) {
	if strings.HasPrefix(urlPath, "//") {
		urlPath = "https" + urlPath
	}
	if !strings.HasPrefix(urlPath, "http") {
		return "", ""
	}
	u, err := url.Parse(urlPath)
	if err != nil {
		return "", ""
	}
	if (u.Port() != "" && u.Port() != "80" && u.Port() != "443") || net.ParseIP(u.Hostname()) != nil {
		return "", ""
	}

	return u.Scheme, u.Hostname()
}