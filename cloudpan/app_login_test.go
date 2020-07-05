package cloudpan

import (
	"fmt"
	"testing"
)

func TestAppLogin(t *testing.T) {
	r, e := AppLogin("131687xxxxx@189.cn", "12345xxxxx")
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Printf("%+v", r)
}
