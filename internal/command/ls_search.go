// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package command

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/library-go/converter"
	"github.com/tickstep/library-go/text"
	"github.com/urfave/cli"
)

type (
	// LsOptions 列目录可选项
	LsOptions struct {
		Total bool
	}

	// SearchOptions 搜索可选项
	SearchOptions struct {
		Total   bool
		Recurse bool
	}
)

const (
	opLs int = iota
	opSearch
)

func CmdLs() cli.Command {
	return cli.Command{
		Name:      "ls",
		Aliases:   []string{"l", "ll"},
		Usage:     "列出目录",
		UsageText: cmder.App().Name + " ls <目录>",
		Description: `
	列出当前工作目录内的文件和目录, 或指定目录内的文件和目录

	示例:

	列出 我的资源 内的文件和目录
	cloudpan189-go ls 我的资源

	绝对路径
	cloudpan189-go ls /我的资源

	降序排序
	cloudpan189-go ls -desc 我的资源

	按文件大小降序排序
	cloudpan189-go ls -size -desc 我的资源

	json 格式输出
	cloudpan189-go ls -json 我的资源
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}
			var (
				orderBy   cloudpan.OrderBy   = cloudpan.OrderByName
				orderSort cloudpan.OrderSort = cloudpan.OrderAsc
			)

			switch {
			case c.IsSet("asc"):
				orderSort = cloudpan.OrderAsc
			case c.IsSet("desc"):
				orderSort = cloudpan.OrderDesc
			default:
				orderSort = cloudpan.OrderAsc
			}

			switch {
			case c.IsSet("time"):
				orderBy = cloudpan.OrderByTime
			case c.IsSet("name"):
				orderBy = cloudpan.OrderByName
			case c.IsSet("size"):
				orderBy = cloudpan.OrderBySize
			default:
				orderBy = cloudpan.OrderByTime
			}
			var json_output bool = false
			if c.IsSet("json") {
				json_output = true
			}

			RunLs(parseFamilyId(c), c.Args().Get(0), &LsOptions{
				Total: c.Bool("l") || c.Parent().Args().Get(0) == "ll",
			}, orderBy, orderSort, json_output)

			return nil
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "l",
				Usage: "详细显示",
			},
			cli.BoolFlag{
				Name:  "asc",
				Usage: "升序排序",
			},
			cli.BoolFlag{
				Name:  "desc",
				Usage: "降序排序",
			},
			cli.BoolFlag{
				Name:  "time",
				Usage: "根据时间排序",
			},
			cli.BoolFlag{
				Name:  "name",
				Usage: "根据文件名排序",
			},
			cli.BoolFlag{
				Name:  "size",
				Usage: "根据大小排序",
			},
			cli.StringFlag{
				Name:  "familyId",
				Usage: "家庭云ID",
				Value: "",
			},
			cli.BoolFlag{
				Name:  "json",
				Usage: "JSON 格式输出",
			},
		},
	}
}

func RunLs(familyId int64, targetPath string, lsOptions *LsOptions, orderBy cloudpan.OrderBy, orderSort cloudpan.OrderSort, isJson bool) {
	activeUser := config.Config.ActiveUser()
	targetPath = activeUser.PathJoin(familyId, targetPath)
	if targetPath[len(targetPath)-1] == '/' {
		targetPath = text.Substr(targetPath, 0, len(targetPath)-1)
	}

	targetPathInfo, err := activeUser.PanClient().AppFileInfoByPath(familyId, targetPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	fileList := cloudpan.AppFileList{}
	fileListParam := cloudpan.NewAppFileListParam()
	fileListParam.FileId = targetPathInfo.FileId
	fileListParam.FamilyId = familyId
	fileListParam.OrderBy = orderBy
	fileListParam.OrderSort = orderSort
	if targetPathInfo.IsFolder {
		fileResult, err := activeUser.PanClient().AppGetAllFileList(fileListParam)
		if err != nil {
			fmt.Println(err)
			return
		}
		fileList = fileResult.FileList

		// more page?
		//if fileResult.RecordCount > fileResult.PageSize {
		//	pageCount := int(math.Ceil(float64(fileResult.RecordCount) / float64(fileResult.PageSize)))
		//	for page := 2; page <= pageCount; page++ {
		//		fileListParam.PageNum = uint(page)
		//		fileResult, err = activeUser.PanClient().FileList(fileListParam)
		//		if err != nil {
		//			fmt.Println(err)
		//			break
		//		}
		//		fileList = append(fileList, fileResult.Data...)
		//	}
		//}
	} else {
		fileList = append(fileList, targetPathInfo)
	}
	if isJson {
		data, err := json.Marshal(fileList)
		if err != nil {
			fmt.Print("{\"error\":\"", err.Error(), "\"}")
		} else {
			fmt.Print(string(data))
		}
	} else {
		renderTable(opLs, lsOptions.Total, targetPath, fileList)
	}
}

func renderTable(op int, isTotal bool, path string, files cloudpan.AppFileList) {
	tb := cmdtable.NewTable(os.Stdout)
	var (
		fN, dN   int64
		showPath string
	)

	switch op {
	case opLs:
		showPath = "文件(目录)"
	case opSearch:
		showPath = "路径"
	}

	if isTotal {
		tb.SetHeader([]string{"#", "file_id", "文件大小", "文件MD5", "文件大小(原始)", "创建日期", "修改日期", showPath})
		tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		for k, file := range files {
			if file.IsFolder {
				tb.Append([]string{strconv.Itoa(k), file.FileId, "-", "-", "-", file.CreateTime, file.LastOpTime, file.FileName + cloudpan.PathSeparator})
				continue
			}

			switch op {
			case opLs:
				tb.Append([]string{strconv.Itoa(k), file.FileId, converter.ConvertFileSize(file.FileSize, 2), file.FileMd5, strconv.FormatInt(file.FileSize, 10), file.CreateTime, file.LastOpTime, file.FileName})
			case opSearch:
				tb.Append([]string{strconv.Itoa(k), file.FileId, converter.ConvertFileSize(file.FileSize, 2), file.FileMd5, strconv.FormatInt(file.FileSize, 10), file.CreateTime, file.LastOpTime, file.Path})
			}
		}
		fN, dN = files.Count()
		tb.Append([]string{"", "", "总: " + converter.ConvertFileSize(files.TotalSize(), 2), "", "", "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})
	} else {
		tb.SetHeader([]string{"#", "文件大小", "修改日期", showPath})
		tb.SetColumnAlignment([]int{tablewriter.ALIGN_DEFAULT, tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		for k, file := range files {
			if file.IsFolder {
				tb.Append([]string{strconv.Itoa(k), "-", file.LastOpTime, file.FileName + cloudpan.PathSeparator})
				continue
			}

			switch op {
			case opLs:
				tb.Append([]string{strconv.Itoa(k), converter.ConvertFileSize(file.FileSize, 2), file.LastOpTime, file.FileName})
			case opSearch:
				tb.Append([]string{strconv.Itoa(k), converter.ConvertFileSize(file.FileSize, 2), file.LastOpTime, file.Path})
			}
		}
		fN, dN = files.Count()
		tb.Append([]string{"", "总: " + converter.ConvertFileSize(files.TotalSize(), 2), "", fmt.Sprintf("文件总数: %d, 目录总数: %d", fN, dN)})
	}

	tb.Render()

	if fN+dN >= 60 {
		fmt.Printf("\n当前目录: %s\n", path)
	}

	fmt.Printf("----\n")
}
