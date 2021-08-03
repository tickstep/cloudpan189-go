// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/urfave/cli"
)

func CmdSign() cli.Command {
	return cli.Command{
		Name:        "sign",
		Usage:       "用户签到",
		Description: "当前帐号进行签到",
		Category:    "天翼云盘账号",
		Before:      cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}
			RunUserSign()
			return nil
		},
	}
}

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
		fmt.Printf("第1次抽奖失败: %s\n", err)
	} else {
		if r.Success {
			fmt.Printf("第1次抽奖成功: %s\n", r.Tip)
		} else {
			fmt.Printf("第1次抽奖失败: %s\n", err)
			return
		}
	}

	r, err = activeUser.PanClient().UserDrawPrize(cloudpan.ActivitySignPhotos)
	if err != nil {
		fmt.Printf("第2次抽奖失败: %s\n", err)
	} else {
		if r.Success {
			fmt.Printf("第2次抽奖成功: %s\n", r.Tip)
		} else {
			fmt.Printf("第2次抽奖失败: %s\n", err)
			return
		}
	}
}
