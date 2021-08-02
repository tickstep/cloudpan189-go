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
	"github.com/olekukonko/tablewriter"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/library-go/converter"
	"github.com/urfave/cli"
	"os"
	"strconv"
)

func CmdRecycle() cli.Command {
	return cli.Command{
		Name:  "recycle",
		Usage: "回收站",
		Description: `
	回收站操作.

	示例:

	1. 从回收站还原两个文件, 其中的两个文件的 file_id 分别为 1013792297798440 和 643596340463870
	cloudpan189-go recycle restore 1013792297798440 643596340463870

	2. 从回收站删除两个文件, 其中的两个文件的 file_id 分别为 1013792297798440 和 643596340463870
	cloudpan189-go recycle delete 1013792297798440 643596340463870

	3. 清空回收站, 程序不会进行二次确认, 谨慎操作!!!
	cloudpan189-go recycle delete -all
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NumFlags() <= 0 || c.NArg() <= 0 {
				cli.ShowCommandHelp(c, c.Command.Name)
			}
			return nil
		},
		Subcommands: []cli.Command{
			{
				Name:      "list",
				Aliases:   []string{"ls", "l"},
				Usage:     "列出回收站文件列表",
				UsageText: cmder.App().Name + " recycle list",
				Action: func(c *cli.Context) error {
					RunRecycleList(c.Int("page"))
					return nil
				},
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "page",
						Usage: "回收站文件列表页数",
						Value: 1,
					},
				},
			},
			{
				Name:        "restore",
				Aliases:     []string{"r"},
				Usage:       "还原回收站文件或目录",
				UsageText:   cmder.App().Name + " recycle restore <file_id 1> <file_id 2> <file_id 3> ...",
				Description: `根据文件/目录的 fs_id, 还原回收站指定的文件或目录`,
				Action: func(c *cli.Context) error {
					if c.NArg() <= 0 {
						cli.ShowCommandHelp(c, c.Command.Name)
						return nil
					}
					RunRecycleRestore(c.Args()...)
					return nil
				},
			},
			{
				Name:        "delete",
				Aliases:     []string{"d"},
				Usage:       "删除回收站文件或目录 / 清空回收站",
				UsageText:   cmder.App().Name + " recycle delete [-all] <file_id 1> <file_id 2> <file_id 3> ...",
				Description: `根据文件/目录的 file_id 或 -all 参数, 删除回收站指定的文件或目录或清空回收站`,
				Action: func(c *cli.Context) error {
					if c.Bool("all") {
						// 清空回收站
						RunRecycleClear()
						return nil
					}

					if c.NArg() <= 0 {
						cli.ShowCommandHelp(c, c.Command.Name)
						return nil
					}
					RunRecycleDelete(c.Args()...)
					return nil
				},
				Flags: []cli.Flag{
					cli.BoolFlag{
						Name:  "all",
						Usage: "清空回收站, 程序不会进行二次确认, 谨慎操作!!!",
					},
				},
			},
		},
	}
}

// RunRecycleList 执行列出回收站文件列表
func RunRecycleList(page int) {
	if page < 1 {
		page = 1
	}

	panClient := GetActivePanClient()
	fdl, err := panClient.RecycleList(page, 0)
	if err != nil {
		fmt.Println(err)
		return
	}

	tb := cmdtable.NewTable(os.Stdout)
	tb.SetHeader([]string{"#", "file_id", "文件名", "文件大小", "创建日期", "修改日期"})
	tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	for k, file := range fdl.Data {
		tb.Append([]string{strconv.Itoa(k), file.FileId, file.FileName, converter.ConvertFileSize(file.FileSize, 2), file.CreateTime, file.LastOpTime})
	}

	tb.Render()
}

// RunRecycleRestore 执行还原回收站文件或目录
func RunRecycleRestore(fidStrList ...string) {
	panClient := GetActivePanClient()
	pageNum := 1

	restoreFileList := []*cloudpan.RecycleFileInfo{}

	fdl, err := panClient.RecycleList(pageNum, 0)
	if err != nil {
		fmt.Printf("还原失败, 请稍后重试")
		return
	}
	isContinue := true
	for {
		if !isContinue || fdl.RecordCount <= 0 {
			break
		}
		pageNum += 1
		for _, f := range fdl.Data {
			if isFileIdInTheRestoreList(f.FileId, fidStrList...) {
				restoreFileList = append(restoreFileList, f)
				if len(restoreFileList) >= len(fidStrList) {
					isContinue = false
					break
				}
			}
		}
		if isContinue {
			fdl, err = panClient.RecycleList(pageNum, 0)
			if err != nil {
				fmt.Printf("还原失败, 请稍后重试")
				return
			}
		}
	}
	if len(restoreFileList) == 0 {
		fmt.Printf("没有需要还原的文件")
		return
	}

	taskId, err := panClient.RecycleRestore(restoreFileList)
	if err != nil {
		fmt.Printf("还原文件失败：%s", err)
		return
	}

	if taskId != "" {
		fmt.Printf("还原成功\n")
	}
}

func isFileIdInTheRestoreList(fileId string, fidStrList ...string) bool {
	for _, id := range fidStrList {
		if id == fileId {
			return true
		}
	}
	return false
}

// RunRecycleDelete 执行删除回收站文件或目录
func RunRecycleDelete(fidStrList ...string) {
	panClient := GetActivePanClient()
	idList := []string{}
	for _, s := range fidStrList {
		idList = append(idList, s)
	}
	err := panClient.RecycleDelete(0, idList)
	if err != nil {
		fmt.Printf("彻底删除文件失败：%s", err)
		return
	}
	fmt.Printf("彻底删除文件成功\n")
}

// RunRecycleClear 清空回收站
func RunRecycleClear() {
	panClient := GetActivePanClient()
	err := panClient.RecycleClear(0)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("清空回收站成功\n")
}
