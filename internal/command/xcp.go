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
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cmder"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/urfave/cli"
)

type (
	FileSourceType string
)

const (
	// 个人云文件
	PersonCloud FileSourceType = "person"

	// 家庭云文件
	FamilyCloud FileSourceType = "family"
)

func CmdXcp() cli.Command {
	return cli.Command{
		Name:  "xcp",
		Usage: "转存拷贝文件/目录，个人云和家庭云之间转存文件",
		UsageText: cmder.App().Name + ` xcp <文件/目录>
	cloudpan189-go xcp <文件/目录1> <文件/目录2> <文件/目录3>`,
		Description: `
	注意: 拷贝多个文件和目录时, 请确保每一个文件和目录都存在, 否则拷贝操作会失败. 同样需要保证目标云不存在对应的文件，否则也会操作失败。

	示例:

	当前程序工作在个人云模式下，将 /个人云目录/1.mp4 转存复制到 家庭云根目录中
	cloudpan189-go xcp /个人云目录/1.mp4

	当前程序工作在家庭云模式下，将 /家庭云目录/1.mp4 和 /家庭云目录/2.mp4 转存复制到 个人云 /来自家庭共享 目录中
	cloudpan189-go xcp /家庭云目录/1.mp4 /家庭云目录/2.mp4
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NArg() <= 0 {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}
			familyId := parseFamilyId(c)
			fileSource := PersonCloud
			if c.IsSet("source") {
				sourceStr := c.String("source")
				if sourceStr == "person" {
					fileSource = PersonCloud
				} else if sourceStr == "family" {
					fileSource = FamilyCloud
				} else {
					fmt.Println("不支持的参数")
					return nil
				}
			} else {
				if IsFamilyCloud(config.Config.ActiveUser().ActiveFamilyId) {
					fileSource = FamilyCloud
				} else {
					fileSource = PersonCloud
				}
			}
			RunXCopy(fileSource, familyId, c.Args()...)
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "familyId",
				Usage:    "家庭云ID",
				Value:    "",
				Required: false,
			},
			cli.StringFlag{
				Name:     "source",
				Usage:    "文件源，person-个人云，family-家庭云",
				Value:    "",
				Required: false,
			},
		},
	}
}

// RunXCopy 执行移动文件/目录
func RunXCopy(source FileSourceType, familyId int64, paths ...string) {
	activeUser := GetActiveUser()

	// use the first family as default
	if familyId == 0 {
		familyResult,err := activeUser.PanClient().AppFamilyGetFamilyList()
		if err != nil {
			fmt.Println("获取家庭列表失败")
			return
		}
		for _,f := range familyResult.FamilyInfoList {
			if f.UserRole == 1 {
				familyId = f.FamilyId
			}
		}
	}

	var opFileList []*cloudpan.AppFileEntity
	var failedPaths []string
	var err error
	switch source {
	case FamilyCloud:
		opFileList, failedPaths, err = GetAppFileInfoByPaths(familyId, paths...)
		break
	case PersonCloud:
		opFileList, failedPaths, err = GetAppFileInfoByPaths(0, paths...)
		break
	default:
		fmt.Println("不支持的云类型")
		return
	}

	if err !=  nil {
		fmt.Println(err)
		return
	}
	if opFileList == nil || len(opFileList) == 0 {
		fmt.Println("没有有效的文件可复制")
		return
	}

	fileIdList := []string{}
	for _,fi := range opFileList {
		fileIdList = append(fileIdList, fi.FileId)
	}

	switch source {
	case FamilyCloud:
		// copy to person cloud
		_,e1 := activeUser.PanClient().AppFamilySaveFileToPersonCloud(familyId, fileIdList)
		if e1 != nil {
			if e1.ErrCode() == apierror.ApiCodeFileAlreadyExisted {
				fmt.Println("复制失败，个人云已经存在对应的文件")
			} else {
				fmt.Println("复制文件到个人云失败")
			}
			return
		}
		break
	case PersonCloud:
		// copy to family cloud
		_,e1 := activeUser.PanClient().AppSaveFileToFamilyCloud(familyId, fileIdList)
		if e1 != nil {
			if e1.ErrCode() == apierror.ApiCodeFileAlreadyExisted {
				fmt.Println("复制失败，家庭云已经存在对应的文件")
			} else {
				fmt.Println("复制文件到家庭云失败")
			}
			return
		}
		break
	default:
		fmt.Println("不支持的云类型")
		return
	}

	if len(failedPaths) > 0 {
		fmt.Println("以下文件复制失败：")
		for _,f := range failedPaths {
			fmt.Println(f)
		}
		fmt.Println("")
	}

	switch source {
	case FamilyCloud:
		// copy to person cloud
		fmt.Println("成功复制以下文件到个人云目录 /来自家庭共享")
		for _,fi := range opFileList {
			fmt.Println(fi.Path)
		}
		break
	case PersonCloud:
		// copy to family cloud
		fmt.Println("成功复制以下文件到家庭云根目录")
		for _,fi := range opFileList {
			fmt.Println(fi.Path)
		}
		break
	default:
		fmt.Println("不支持的云类型")
		return
	}
}