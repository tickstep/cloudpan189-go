package getip

import (
	"bytes"
	"github.com/tickstep/cloudpan189-go/library/requester"
	"net/http"
	"unsafe"
)

func IPInfoFromTechainBaiduByClient(c *requester.HTTPClient) (ipAddr string, err error) {
	body, err := c.Fetch(http.MethodGet, "https://techain.baidu.com/srcmon", nil, map[string]string{
		"User-Agent":      "x18/600000101/10.0.63/4.1.3",
		"Pragma":          "no-cache",
		"Accept":          "*/*",
		"Content-Type":    "application/x-www-form-urlencoded",
		"x-auth-ver":      "1",
		"Accept-Language": "zh-CN",
		"x-device-id":     "00000000000000000000000000000000",
	})
	if err != nil {
		return
	}
	p := bytes.TrimSpace(body)
	return *(*string)(unsafe.Pointer(&p)), nil
}

// IPInfoFromTechainBaidu 从 techain.baidu.com 获取ip
func IPInfoFromTechainBaidu() (ipAddr string, err error) {
	c := requester.NewHTTPClient()
	return IPInfoFromNeteaseByClient(c)
}
