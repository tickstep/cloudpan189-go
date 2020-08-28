package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
)

func RunUserSign() {
	activeUser := GetActiveUser()
	result, err := activeUser.PanClient().AppUserSign()
	if err != nil {
		fmt.Printf("签到失败: %s\n", err)
		return
	}
	if result.Status == cloudpan.AppUserSignStatusSuccess {
		fmt.Printf("签到成功，%s\n", result.Tip)
	} else if result.Status == cloudpan.AppUserSignStatusHasSign {
		fmt.Printf("今日已签到，%s\n", result.Tip)
	} else {
		fmt.Printf("签到失败，%s\n", result.Tip)
	}

	// 抽奖
	r, err := activeUser.PanClient().UserDrawPrize(cloudpan.ActivitySignin)
	if err != nil {
		fmt.Printf("抽奖失败: %s\n", err)
		return
	}
	if r.Success {
		fmt.Printf("抽奖成功: %s\n", r.Tip)
	} else {
		fmt.Printf("抽奖失败: %s\n", err)
		return
	}

	r, err = activeUser.PanClient().UserDrawPrize(cloudpan.ActivitySignPhotos)
	if err != nil {
		fmt.Printf("抽奖失败: %s\n", err)
		return
	}
	if r.Success {
		fmt.Printf("抽奖成功: %s\n", r.Tip)
	} else {
		fmt.Printf("抽奖失败: %s\n", err)
		return
	}
}
