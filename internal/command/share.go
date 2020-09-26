package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/library-go/text"
	"os"
	"strconv"
	"strings"
)

// RunShareSet 执行分享
func RunShareSet(paths []string, expiredTime cloudpan.ShareExpiredTime) {
	fileList, _, err := GetFileInfoByPaths(paths[:len(paths)]...)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, fi := range fileList {
		r, err := GetActivePanClient().SharePrivate(fi.FileId, expiredTime)
		if err != nil {
			fmt.Printf("创建分享链接失败: %s - %s\n", fi.Path, err)
			continue
		}
		fmt.Printf("路径: %s 链接: %s（访问码：%s）\n",fi.Path, r.ShortShareUrl, r.AccessCode)
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
		tb.Append([]string{strconv.Itoa(k), strconv.FormatInt(record.ShareId, 10), record.AccessURL, record.AccessCode, record.FileName, record.FileId, record.ShareTime})
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
	if savePanDirPath[len(savePanDirPath) - 1] == '/' {
		savePanDirPath = text.Substr(savePanDirPath, 0, len(savePanDirPath) - 1)
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