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
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/library-go/text"
)

// RunShareSet 执行分享
func RunShareSet(paths []string, expiredTime cloudpan.ShareExpiredTime, shareMode cloudpan.ShareMode) {
	fileList, _, err := GetFileInfoByPaths(paths[:len(paths)]...)
	if err != nil {
		fmt.Println(err)
		return
	}

	if shareMode == 1 {
		for _, fi := range fileList {
			r, err := GetActivePanClient().SharePrivate(fi.FileId, expiredTime)
			if err != nil {
				fmt.Printf("创建分享链接失败: %s - %s\n", fi.Path, err)
				continue
			}
			fmt.Printf("路径: %s\n链接: %s（访问码：%s）\n", fi.Path, r.ShortShareUrl, r.AccessCode)
		}
	} else {
		for _, fi := range fileList {
			r, err := GetActivePanClient().SharePublic(fi.FileId, expiredTime)
			if err != nil {
				fmt.Printf("创建分享链接失败: %s - %s\n", fi.Path, err)
				continue
			}
			fmt.Printf("路径: %s\n链接: %s\n", fi.Path, r.ShortShareUrl)
		}
	}

}

// RunShareList 执行列出分享列表
func RunShareList(page int) {
	if page < 1 {
		page = 1
	}

	activeUser := GetActiveUser()
	param := cloudpan.NewShareListParam()
	param.PageNum = page
	records, err := activeUser.PanClient().ShareList(param)
	if err != nil {
		fmt.Printf("获取分享列表失败: %s\n", err)
		return
	}

	tb := cmdtable.NewTable(os.Stdout)
	tb.SetHeader([]string{"#", "ShARE_ID", "分享链接", "访问码", "文件名", "FILE_ID", "分享时间"})
	for k, record := range records.Data {
		tm := time.Unix(record.ShareTime / 1000, 0)
		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(record.ShareId, 10), record.AccessURL, record.AccessCode, record.FileName, record.FileId, tm.Format("2006-01-02 15:04:05")})
	}
	tb.Render()
}

// RunShareCancel 执行取消分享
func RunShareCancel(shareIDs []int64) {
	if len(shareIDs) == 0 {
		fmt.Printf("取消分享操作失败, 没有任何 shareid\n")
		return
	}

	activeUser := GetActiveUser()
	b, err := activeUser.PanClient().ShareCancel(shareIDs)
	if err != nil {
		fmt.Printf("取消分享操作失败: %s\n", err)
		return
	}

	if b {
		fmt.Printf("取消分享操作成功\n")
	} else {
		fmt.Printf("取消分享操作失败\n")
	}
}

func RunShareSave(shareUrl, savePanDirPath string) {
	activeUser := GetActiveUser()

	if shareUrl == "" || strings.Index(shareUrl, cloudpan.WEB_URL) < 0 {
		fmt.Printf("分享链接错误\n")
		return
	}
	if savePanDirPath == "" {
		fmt.Printf("指定的网盘文件夹路径有误\n")
		return
	}

	shareUrl = strings.ReplaceAll(shareUrl, "（访问码：", " ")
	shareUrl = strings.ReplaceAll(shareUrl, "）", "")

	idxBlank := strings.Index(shareUrl, " ")

	if idxBlank < 0 {
		fmt.Printf("分享链接错误\n")
		return
	}

	accessUrl := strings.Trim(shareUrl[:idxBlank], " ")
	accessCode := strings.Trim(shareUrl[idxBlank+1:], " ")

	if accessUrl == "" || accessCode == "" {
		fmt.Printf("分享链接提取错误\n")
		return
	}

	savePanDirPath = activeUser.PathJoin(0, savePanDirPath)
	if savePanDirPath[len(savePanDirPath)-1] == '/' {
		savePanDirPath = text.Substr(savePanDirPath, 0, len(savePanDirPath)-1)
	}
	fi, apier := activeUser.PanClient().FileInfoByPath(savePanDirPath)
	if apier != nil {
		fmt.Printf("指定的网盘文件夹路径有误\n")
		return
	}
	if fi == nil || !fi.IsFolder {
		fmt.Printf("指定的网盘路径不是文件夹\n")
		return
	}

	b, apier := activeUser.PanClient().ShareSave(accessUrl, accessCode, fi.FileId)
	if apier != nil || !b {
		fmt.Printf("转存出错：%s\n", apier)
		return
	}
	fmt.Printf("转存成功\n")
}
