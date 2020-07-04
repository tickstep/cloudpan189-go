package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"os"
	"strconv"
	"time"
)

// RunRemove 执行 批量删除文件/目录
func RunRemove(paths ...string) {
	activeUser := GetActiveUser()
	infoList, _, delFileInfos := getBatchTaskInfoList(paths...)
	if infoList == nil || len(*infoList) == 0 {
		fmt.Println("没有有效的文件可删除")
		return
	}

	// create delete files task
	delParam := &cloudpan.BatchTaskParam{
		TypeFlag: cloudpan.BatchTaskTypeDelete,
		TaskInfos: *infoList,
	}

	taskId, err := activeUser.PanClient().CreateBatchTask(delParam)
	if err != nil {
		fmt.Println("无法删除文件，请稍后重试")
		return
	}
	logger.Verboseln("delete file task id: " + taskId)

	// check task
	time.Sleep(time.Duration(200) * time.Millisecond)
	taskRes, err := activeUser.PanClient().CheckBatchTask(cloudpan.BatchTaskTypeDelete, taskId)
	if err != nil || taskRes.TaskStatus != cloudpan.BatchTaskStatusOk {
		fmt.Println("无法删除文件，请稍后重试")
		return
	}

	pnt := func() {
		tb := cmdtable.NewTable(os.Stdout)
		tb.SetHeader([]string{"#", "文件/目录"})
		for k := range *delFileInfos {
			tb.Append([]string{strconv.Itoa(k), (*delFileInfos)[k].Path})
		}
		tb.Render()
	}
	if taskRes.TaskStatus == cloudpan.BatchTaskStatusOk {
		fmt.Println("操作成功, 以下文件/目录已删除, 可在云盘文件回收站找回: ")
		pnt()
	}
}

func getBatchTaskInfoList(paths ...string) (*cloudpan.BatchTaskInfoList, *[]string, *[]*cloudpan.FileEntity) {
	activeUser := GetActiveUser()
	failedRmPaths := make([]string, 0, len(paths))
	delFileInfos := make([]*cloudpan.FileEntity, 0, len(paths))
	infoList := cloudpan.BatchTaskInfoList{}
	for _, p := range paths {
		absolutePath := activeUser.PathJoin(p)
		fe, err := activeUser.PanClient().FileInfoByPath(absolutePath)
		if err != nil {
			failedRmPaths = append(failedRmPaths, absolutePath)
			continue
		}
		isFolder := 0
		if fe.IsFolder {
			isFolder = 1
		}
		infoItem := &cloudpan.BatchTaskInfo{
			FileId: fe.FileId,
			FileName: fe.FileName,
			IsFolder: isFolder,
			SrcParentId: fe.ParentId,
		}
		infoList = append(infoList, infoItem)
		delFileInfos = append(delFileInfos, fe)
	}
	return &infoList, &failedRmPaths, &delFileInfos
}