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
	"github.com/tickstep/cloudpan189-api/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/cmder"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/urfave/cli"
	"path"
	"strings"
)

func CmdRename() cli.Command {
	return cli.Command{
		Name:  "rename",
		Usage: "重命名文件",
		UsageText: `重命名文件:
	cloudpan189-go rename <旧文件/目录名> <新文件/目录名>`,
		Description: `
	示例:

	将文件 1.mp4 重命名为 2.mp4
	cloudpan189-go rename 1.mp4 2.mp4

	将文件 /test/1.mp4 重命名为 /test/2.mp4
	要求必须是同一个文件目录内
	cloudpan189-go rename /test/1.mp4 /test/2.mp4
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NArg() != 2 {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}
			RunRename(parseFamilyId(c), c.Args().Get(0), c.Args().Get(1))
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

func RunRename(familyId int64, oldName string, newName string) {
	if oldName == "" {
		fmt.Println("请指定命名文件")
		return
	}
	if newName == "" {
		fmt.Println("请指定文件新名称")
		return
	}
	activeUser := GetActiveUser()
	oldName = activeUser.PathJoin(familyId, strings.TrimSpace(oldName))
	newName = activeUser.PathJoin(familyId, strings.TrimSpace(newName))
	if path.Dir(oldName) != path.Dir(newName) {
		fmt.Println("只能命名同一个目录的文件")
		return
	}
	if !apiutil.CheckFileNameValid(path.Base(newName)) {
		fmt.Println("文件名不能包含特殊字符：" + apiutil.FileNameSpecialChars)
		return
	}

	fileId := ""
	r, err := GetActivePanClient().AppFileInfoByPath(familyId, activeUser.PathJoin(familyId, oldName))
	if err != nil {
		fmt.Printf("原文件不存在： %s, %s\n", oldName, err)
		return
	}
	fileId = r.FileId

	var b *cloudpan.AppFileEntity
	var e *apierror.ApiError
	if IsFamilyCloud(familyId) {
		b, e = activeUser.PanClient().AppFamilyRenameFile(familyId, fileId, path.Base(newName))
	} else {
		b, e = activeUser.PanClient().AppRenameFile(fileId, path.Base(newName))
	}
	if e != nil {
		fmt.Println(e.Err)
		return
	}
	if b == nil {
		fmt.Println("重命名文件失败")
		return
	}
	fmt.Printf("重命名文件成功：%s -> %s\n", path.Base(oldName), path.Base(newName))
}
