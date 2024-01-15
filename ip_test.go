package cobweb

import (
	"log"
	"testing"
)

func TestCheckIps(t *testing.T) {
	CheckIps()
}

func TestPing(t *testing.T) {
	ip := "124.70.135.2"

	res := Ping(ip, "80")
	log.Println(res)
}

func TestGetAllCSegments(t *testing.T) {
	GetAllCSegments()
}
