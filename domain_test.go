package cobweb

import "testing"

func TestGetDomains(t *testing.T) {
	getDomains("117.21.178.0")
}

func TestGetDomainByIp(t *testing.T) {
	getDomainByIp("https://ipchaxun.com/117.21.178.232/", "https://ipchaxun.com/117.21.178.232/")
}

func TestGetContents(t *testing.T) {
	GetContents()
}
