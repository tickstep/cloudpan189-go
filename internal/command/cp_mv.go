package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"path"
	"strconv"
	"time"
)

// RunCopy 执行复制文件/目录
func RunCopy(paths ...string) {
	activeUser := GetActiveUser()
	opFileList, targetFile, _, err := getFileInfo(paths...)
	if err !=  nil {
		fmt.Println(err)
		return
	}
	if targetFile == nil {
		fmt.Println("目标文件不存在")
		return
	}
	if opFileList == nil || len(opFileList) == 0 {
		fmt.Println("没有有效的文件可复制")
		return
	}

	// create task
	targetFolderId, _ := strconv.Atoi(targetFile.FileId)
	taskParam := &cloudpan.BatchTaskParam{
		TypeFlag: cloudpan.BatchTaskTypeCopy,
		TaskInfos: makeBatchTaskInfoList(opFileList),
		TargetFolderId: targetFolderId,
	}

	taskId, err1 := activeUser.PanClient().CreateBatchTask(taskParam)
	if err1 != nil {
		fmt.Println("无法复制文件，请稍后重试")
		return
	}
	logger.Verboseln("task id: " + taskId)

	// check task
	time.Sleep(time.Duration(200) * time.Millisecond)
	taskRes, err2 := activeUser.PanClient().CheckBatchTask(cloudpan.BatchTaskTypeCopy, taskId)
	if err2 != nil {
		fmt.Println("无法复制文件，请稍后重试")
		return
	}
	if taskRes.TaskStatus == cloudpan.BatchTaskStatusNotAction {
		fmt.Println("无法复制文件，文件可能已经存在")
		return
	}

	if taskRes.TaskStatus == cloudpan.BatchTaskStatusOk {
		fmt.Println("操作成功, 已复制文件到目标目录: ", targetFile.Path)
	}
}


// RunMove 执行移动文件/目录
func RunMove(paths ...string) {
	activeUser := GetActiveUser()
	opFileList, targetFile, _, err := getFileInfo(paths...)
	if err !=  nil {
		fmt.Println(err)
		return
	}
	if targetFile == nil {
		fmt.Println("目标文件不存在")
		return
	}
	if opFileList == nil || len(opFileList) == 0 {
		fmt.Println("没有有效的文件可移动")
		return
	}

	// create task
	targetFolderId, _ := strconv.Atoi(targetFile.FileId)
	taskParam := &cloudpan.BatchTaskParam{
		TypeFlag: cloudpan.BatchTaskTypeMove,
		TaskInfos: makeBatchTaskInfoList(opFileList),
		TargetFolderId: targetFolderId,
	}

	taskId, err1 := activeUser.PanClient().CreateBatchTask(taskParam)
	if err1 != nil {
		fmt.Println("无法移动文件，请稍后重试")
		return
	}
	logger.Verboseln("task id: " + taskId)

	// check task
	time.Sleep(time.Duration(200) * time.Millisecond)
	taskRes, err2 := activeUser.PanClient().CheckBatchTask(cloudpan.BatchTaskTypeMove, taskId)
	if err2 != nil {
		fmt.Println("无法移动文件，请稍后重试")
		return
	}
	if taskRes.TaskStatus == cloudpan.BatchTaskStatusNotAction {
		fmt.Println("无法移动文件，文件可能已经存在")
		return
	}

	if taskRes.TaskStatus == cloudpan.BatchTaskStatusOk {
		fmt.Println("操作成功, 已移动文件到目标目录: ", targetFile.Path)
	}
}

func getFileInfo(paths ...string) (opFileList []*cloudpan.FileEntity, targetFile *cloudpan.FileEntity, failedPaths []string, error error) {
	if len(paths) <= 1 {
		return nil, nil, nil, fmt.Errorf("请指定目标文件夹路径")
	}
	activeUser := GetActiveUser()
	// the last one is the target file path
	targetFilePath := path.Clean(paths[len(paths)-1])
	absolutePath := activeUser.PathJoin(targetFilePath)
	targetFile, err := activeUser.PanClient().FileInfoByPath(absolutePath)
	if err != nil || !targetFile.IsFolder {
		return nil, nil, nil, fmt.Errorf("指定目标文件夹不存在")
	}

	for idx := 0; idx < (len(paths)-1); idx++ {
		absolutePath := path.Clean(activeUser.PathJoin(paths[idx]))
		fe, err := activeUser.PanClient().FileInfoByPath(absolutePath)
		if err != nil {
			failedPaths = append(failedPaths, absolutePath)
			continue
		}
		opFileList = append(opFileList, fe)
	}
	return
}

func makeBatchTaskInfoList(opFileList []*cloudpan.FileEntity) (infoList cloudpan.BatchTaskInfoList) {
	for _, fe := range opFileList {
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
	}
	return
}