package command

import (
	"fmt"
)

func RunUserSign() {
	activeUser := GetActiveUser()
	r, err := activeUser.PanClient().AppUserSign(&activeUser.AppToken)
	if err != nil {
		fmt.Printf("签到失败: %s\n", err)
		return
	}
	fmt.Println(r)
}
