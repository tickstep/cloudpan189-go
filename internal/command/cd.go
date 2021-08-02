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
	"github.com/tickstep/cloudpan189-go/cmder"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/urfave/cli"
)

func CmdCd() cli.Command {
	return cli.Command{
		Name:     "cd",
		Category: "天翼云盘",
		Usage:    "切换工作目录",
		Description: `
	cloudpan189-go cd <目录, 绝对路径或相对路径>

	示例:

	切换 /我的资源 工作目录:
	cloudpan189-go cd /我的资源

	切换上级目录:
	cloudpan189-go cd ..

	切换根目录:
	cloudpan189-go cd /
`,
		Before: cmder.ReloadConfigFunc,
		After:  cmder.SaveConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NArg() == 0 {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}
			RunChangeDirectory(parseFamilyId(c), c.Args().Get(0))
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "familyId",
				Usage: "家庭云ID",
				Value: "",
			},
		},
	}
}

func CmdPwd() cli.Command {
	return cli.Command{
		Name:      "pwd",
		Usage:     "输出工作目录",
		UsageText: cmder.App().Name + " pwd",
		Category:  "天翼云盘",
		Before:    cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}
			if IsFamilyCloud(config.Config.ActiveUser().ActiveFamilyId) {
				fmt.Println(config.Config.ActiveUser().FamilyWorkdir)
			} else {
				fmt.Println(config.Config.ActiveUser().Workdir)
			}
			return nil
		},
	}
}

func RunChangeDirectory(familyId int64, targetPath string) {
	user := config.Config.ActiveUser()
	targetPath = user.PathJoin(familyId, targetPath)

	targetPathInfo, err := user.PanClient().AppFileInfoByPath(familyId, targetPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !targetPathInfo.IsFolder {
		fmt.Printf("错误: %s 不是一个目录 (文件夹)\n", targetPath)
		return
	}

	if IsFamilyCloud(familyId) {
		user.FamilyWorkdir = targetPath
		user.FamilyWorkdirFileEntity = *targetPathInfo
	} else {
		user.Workdir = targetPath
		user.WorkdirFileEntity = *targetPathInfo
	}

	fmt.Printf("改变工作目录: %s\n", targetPath)
}
