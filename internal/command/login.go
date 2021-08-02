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
	_ "github.com/tickstep/library-go/requester"
	"github.com/urfave/cli"
)


func CmdLogin() cli.Command {
	return cli.Command{
		Name:  "login",
		Usage: "登录天翼云盘账号",
		Description: `
	示例:
		cloudpan189-go login
		cloudpan189-go login -username=tickstep -password=123xxx

	常规登录:
		按提示一步一步来即可.
`,
		Category: "天翼云盘账号",
		Before:   cmder.ReloadConfigFunc, // 每次进行登录动作的时候需要调用刷新配置
		After:    cmder.SaveConfigFunc, // 登录完成需要调用保存配置
		Action: func(c *cli.Context) error {
			appToken := cloudpan.AppLoginToken{}
			webToken := cloudpan.WebLoginToken{}
			username := ""
			passowrd := ""
			if c.IsSet("COOKIE_LOGIN_USER") {
				webToken.CookieLoginUser = c.String("COOKIE_LOGIN_USER")
			} else if c.NArg() == 0 {
				var err error
				username, passowrd, webToken, appToken, err = RunLogin(c.String("username"), c.String("password"))
				if err != nil {
					fmt.Println(err)
					return err
				}
			} else {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}
			cloudUser, _ := config.SetupUserByCookie(&webToken, &appToken)
			// save username / password
			cloudUser.LoginUserName = config.EncryptString(username)
			cloudUser.LoginUserPassword = config.EncryptString(passowrd)
			config.Config.SetActiveUser(cloudUser)
			fmt.Println("天翼帐号登录成功: ", cloudUser.Nickname)
			return nil
		},
		// 命令的附加options参数说明，使用 help login 命令即可查看
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "username",
				Usage: "登录天翼帐号的用户名(手机号/邮箱/别名)",
			},
			cli.StringFlag{
				Name:  "password",
				Usage: "登录天翼帐号的用户密码",
			},
			// 暂不支持
			// cloudpan189-go login -COOKIE_LOGIN_USER=8B12CBBCE89CA8DFC3445985B63B511B5E7EC7...
			//cli.StringFlag{
			//	Name:  "COOKIE_LOGIN_USER",
			//	Usage: "使用 COOKIE_LOGIN_USER cookie来登录帐号",
			//},
		},
	}
}

func CmdLogout() cli.Command {
	return cli.Command{
		Name:        "logout",
		Usage:       "退出天翼帐号",
		Description: "退出当前登录的帐号",
		Category:    "天翼云盘账号",
		Before:      cmder.ReloadConfigFunc,
		After:       cmder.SaveConfigFunc,
		Action: func(c *cli.Context) error {
			if config.Config.NumLogins() == 0 {
				fmt.Println("未设置任何帐号, 不能退出")
				return nil
			}

			var (
				confirm    string
				activeUser = config.Config.ActiveUser()
			)

			if !c.Bool("y") {
				fmt.Printf("确认退出当前帐号: %s ? (y/n) > ", activeUser.Nickname)
				_, err := fmt.Scanln(&confirm)
				if err != nil || (confirm != "y" && confirm != "Y") {
					return err
				}
			}

			deletedUser, err := config.Config.DeleteUser(activeUser.UID)
			if err != nil {
				fmt.Printf("退出用户 %s, 失败, 错误: %s\n", activeUser.Nickname, err)
			}

			fmt.Printf("退出用户成功: %s\n", deletedUser.Nickname)
			return nil
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "y",
				Usage: "确认退出帐号",
			},
		},
	}
}

func RunLogin(username, password string) (usernameStr, passwordStr string, webToken cloudpan.WebLoginToken, appToken cloudpan.AppLoginToken, error error) {
	return cmder.DoLoginHelper(username, password)
}
