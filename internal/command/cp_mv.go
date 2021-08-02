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
	"github.com/tickstep/cloudpan189-go/cmder"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/library-go/logger"
	"github.com/urfave/cli"
	"path"
	"time"
)

func CmdCp() cli.Command {
	return cli.Command{
		Name:  "cp",
		Usage: "拷贝文件/目录",
		UsageText: cmder.App().Name + ` cp <文件/目录> <目标文件/目录>
	cloudpan189-go cp <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>`,
		Description: `
	注意: 拷贝多个文件和目录时, 请确保每一个文件和目录都存在, 否则拷贝操作会失败.

	示例:

	将 /我的资源/1.mp4 复制到 根目录 /
	cloudpan189-go cp /我的资源/1.mp4 /

	将 /我的资源/1.mp4 和 /我的资源/2.mp4 复制到 根目录 /
	cloudpan189-go cp /我的资源/1.mp4 /我的资源/2.mp4 /
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NArg() <= 1 {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}
			if IsFamilyCloud(config.Config.ActiveUser().ActiveFamilyId) {
				fmt.Println("家庭云不支持复制操作")
				return nil
			}
			RunCopy(c.Args()...)
			return nil
		},
	}
}

func CmdMv() cli.Command {
	return cli.Command{
		Name:  "mv",
		Usage: "移动文件/目录",
		UsageText: `移动:
	cloudpan189-go mv <文件/目录1> <文件/目录2> <文件/目录3> ... <目标目录>`,
		Description: `
	注意: 移动多个文件和目录时, 请确保每一个文件和目录都存在, 否则移动操作会失败.

	示例:

	将 /我的资源/1.mp4 移动到 根目录 /
	cloudpan189-go mv /我的资源/1.mp4 /
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NArg() <= 1 {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}
			if config.Config.ActiveUser() == nil {
				fmt.Println("未登录账号")
				return nil
			}

			RunMove(parseFamilyId(c), c.Args()...)
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