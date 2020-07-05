package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"os"
	"strconv"
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