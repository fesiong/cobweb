package cobweb

import (
	"testing"
)

// Test for Request.
func TestRequest(t *testing.T) {
	link := "http://jh.517sichuan.com/"
	resp, err := Request(link)
	if err != nil {
		t.Error(err.Error())
	}
	t.Skip(resp)
}