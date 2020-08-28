package command

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/library-go/converter"
	"os"
	"strconv"
)

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
