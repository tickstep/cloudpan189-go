package cloudpan

import (
	"fmt"
	"testing"
)

func TestLogin(t *testing.T) {
	Login("12345@189.cn", "123")
}

func TestGetCaptchaImage(t *testing.T) {
	s, e := GetCaptchaImage()
	fmt.Println(s)
	fmt.Println(e)
}
