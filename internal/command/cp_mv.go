package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/library-go/logger"
	"path"
	"time"
)

// RunCopy 执行复制文件/目录
func RunCopy(paths ...string) {
	// 只支持个人云
	familyId := int64(0)
	activeUser := GetActiveUser()
	opFileList, targetFile, _, err := getFileInfo(familyId, paths...)
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
	taskParam := &cloudpan.BatchTaskParam{
		TypeFlag: cloudpan.BatchTaskTypeCopy,
		TaskInfos: makeBatchTaskInfoList(opFileList),
		TargetFolderId: targetFile.FileId,
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
func RunMove(familyId int64, paths ...string) {
	activeUser := GetActiveUser()
	opFileList, targetFile, _, err := getFileInfo(familyId, paths...)
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

	if IsFamilyCloud(familyId) {
		failedMoveFiles := []*cloudpan.AppFileEntity{}
		b := false
		for _,mfi := range opFileList {
			_,er := activeUser.PanClient().AppFamilyMoveFile(familyId, mfi.FileId, targetFile.FileId)
			if er != nil {
				failedMoveFiles = append(failedMoveFiles, mfi)
			}
			b = true
		}
		if len(failedMoveFiles) > 0 {
			fmt.Println("以下文件移动失败：")
			for _,f := range failedMoveFiles {
				fmt.Println(f.FileName)
			}
			fmt.Println("")
		}
		if b {
			fmt.Println("操作成功, 已移动文件到目标目录: ", targetFile.Path)
		} else {
			fmt.Println("无法移动文件，请稍后重试")
		}
	} else {
		// create task
		taskParam := &cloudpan.BatchTaskParam{
			TypeFlag: cloudpan.BatchTaskTypeMove,
			TaskInfos: makeBatchTaskInfoList(opFileList),
			TargetFolderId: targetFile.FileId,
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
}

func getFileInfo(familyId int64, paths ...string) (opFileList []*cloudpan.AppFileEntity, targetFile *cloudpan.AppFileEntity, failedPaths []string, error error) {
	if len(paths) <= 1 {
		return nil, nil, nil, fmt.Errorf("请指定目标文件夹路径")
	}
	activeUser := GetActiveUser()
	// the last one is the target file path
	targetFilePath := path.Clean(paths[len(paths)-1])
	absolutePath := activeUser.PathJoin(familyId, targetFilePath)
	targetFile, err := activeUser.PanClient().AppFileInfoByPath(familyId, absolutePath)
	if err != nil || !targetFile.IsFolder {
		return nil, nil, nil, fmt.Errorf("指定目标文件夹不存在")
	}

	opFileList, failedPaths, error = GetAppFileInfoByPaths(familyId, paths[:len(paths)-1]...)
	return
}

func makeBatchTaskInfoList(opFileList []*cloudpan.AppFileEntity) (infoList cloudpan.BatchTaskInfoList) {
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