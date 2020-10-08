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
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/library-go/logger"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

type (
	dirFileListData struct {
		Dir *cloudpan.AppMkdirResult
		File *cloudpan.AppFileListResult
	}
)

const (
	DefaultSaveToPanPath = "/cloudpan189-go"
)

func RunImportFiles(familyId int64, overwrite bool, panSavePath, localFilePath string) {
	lfi,_ := os.Stat(localFilePath)
	if lfi != nil {
		if lfi.IsDir() {
			fmt.Println("请指定导入文件")
			return
		}
	} else {
		// create file
		fmt.Println("导入文件不存在")
		return
	}

	if panSavePath == "" {
		// use default
		panSavePath = DefaultSaveToPanPath
	}

	fmt.Println("导入的文件会存储到目录：" + panSavePath)

	importFile, err := os.OpenFile(localFilePath, os.O_RDONLY, 0755)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer importFile.Close()

	fileData,err := ioutil.ReadAll(importFile)
	if err != nil {
		fmt.Println("读取文件出错")
		return
	}
	fileText := string(fileData)
	if len(fileText) == 0 {
		fmt.Println("文件为空")
		return
	}
	fileText = strings.TrimSpace(fileText)
	fileLines := strings.Split(fileText, "\n")
	importFileItems := []ImportExportFileItem{}
	for _,line := range fileLines {
		line = strings.TrimSpace(line)
		item := &ImportExportFileItem{}
		if err := json.Unmarshal([]byte(line), item); err != nil {
			logger.Verboseln("parse line failed: " + line)
			fmt.Println("Error Data: " + line)
			continue
		}
		item.Path = path.Join(panSavePath, item.Path)
		importFileItems = append(importFileItems, *item)
	}
	if len(importFileItems) == 0 {
		fmt.Println("没有可以导入的文件项目")
		return
	}

	fmt.Println("正在准备导入...")
	dirMap := prepareMkdir(familyId, importFileItems)

	fmt.Println("正在导入...")
	successImportFiles := []ImportExportFileItem{}
	failedImportFiles := []ImportExportFileItem{}
	for _,item := range importFileItems {
		fmt.Printf("正在处理导入: %s\n", item.Path)
		result, abort := processOneImport(familyId, overwrite, dirMap, item)
		if abort {
			fmt.Println("导入任务终止了")
			break
		}
		if result {
			successImportFiles = append(successImportFiles, item)
		} else {
			failedImportFiles = append(failedImportFiles, item)
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	if len(failedImportFiles) > 0 {
		fmt.Println("\n以下文件导入失败")
		for _,f := range failedImportFiles {
			fmt.Printf("%s %s\n", f.FileMd5, f.Path)
		}
		fmt.Println("")
	}
	fmt.Printf("导入结果, 成功 %d, 失败 %d\n", len(successImportFiles), len(failedImportFiles))
}

func processOneImport(familyId int64, isOverwrite bool, dirMap map[string]*dirFileListData, item ImportExportFileItem) (result, abort bool) {
	panClient := config.Config.ActiveUser().PanClient()
	panDir,fileName := path.Split(item.Path)
	dataItem := dirMap[path.Dir(panDir)]
	if isOverwrite {
		// 标记覆盖旧同名文件
		// 检查同名文件是否存在
		var efi *cloudpan.AppFileEntity = nil
		for _,fileItem := range dataItem.File.FileList {
			if !fileItem.IsFolder && fileItem.FileName == fileName {
				efi = fileItem
				break
			}
		}
		if efi != nil && efi.FileId != "" {
			// existed, delete it
			infoList := cloudpan.BatchTaskInfoList{}
			isFolder := 0
			if efi.IsFolder {
				isFolder = 1
			}
			infoItem := &cloudpan.BatchTaskInfo{
				FileId: efi.FileId,
				FileName: efi.FileName,
				IsFolder: isFolder,
				SrcParentId: efi.ParentId,
			}
			infoList = append(infoList, infoItem)
			delParam := &cloudpan.BatchTaskParam{
				TypeFlag: cloudpan.BatchTaskTypeDelete,
				TaskInfos: infoList,
			}

			var taskId string
			var err *apierror.ApiError
			if familyId > 0 {
				taskId, err = panClient.AppCreateBatchTask(familyId, delParam)
			} else {
				taskId, err = panClient.CreateBatchTask(delParam)
			}

			if err != nil || taskId == "" {
				fmt.Println("无法删除文件，请稍后重试")
				return false, false
			}
			time.Sleep(time.Duration(500) * time.Millisecond)
			fmt.Println("检测到同名文件，已移动到回收站")
		}
	}

	var r *cloudpan.AppCreateUploadFileResult
	var apierr *apierror.ApiError
	ts := time.Now().Format("2006-01-02 15:04:05")
	if item.LastOpTime == "" {
		ts = item.LastOpTime
	}
	appCreateUploadFileParam := &cloudpan.AppCreateUploadFileParam{
		ParentFolderId: dataItem.Dir.FileId,
		FileName: fileName,
		Size: item.FileSize,
		Md5: strings.ToUpper(item.FileMd5),
		LastWrite: ts,
		LocalPath: "",
		FamilyId: familyId,
	}
	if familyId > 0 {
		r, apierr = panClient.AppFamilyCreateUploadFile(appCreateUploadFileParam)
	} else {
		r, apierr = panClient.AppCreateUploadFile(appCreateUploadFileParam)
	}
	if apierr != nil {
		if apierr.Code == apierror.UserDayFlowOverLimited {
			fmt.Println("创建上传任务失败: " + apierr.Error())
			return false, true
		}
		fmt.Println("创建上传任务失败")
		return false, true
	}

	if r.FileDataExists == 1 {
		var er *apierror.ApiError
		if familyId > 0 {
			_, er = panClient.AppFamilyUploadFileCommit(familyId, r.FileCommitUrl, r.UploadFileId, r.XRequestId)
		} else {
			_, er = panClient.AppUploadFileCommit(r.FileCommitUrl, r.UploadFileId, r.XRequestId)
		}
		if er != nil {
			fmt.Println("秒传失败")
			return false, false
		} else {
			return true, false
		}
	} else {
		fmt.Printf("文件未曾上传，无法秒传\n")
		return false, false
	}
}

func prepareMkdir(familyId int64, importFileItems []ImportExportFileItem) map[string]*dirFileListData {
	panClient := config.Config.ActiveUser().PanClient()
	resultMap := map[string]*dirFileListData{}
	for _,item := range importFileItems {
		var apierr *apierror.ApiError
		var rs *cloudpan.AppMkdirResult
		panDir := path.Dir(item.Path)
		if resultMap[panDir] != nil {
			continue
		}
		if panDir != "/" {
			rs, apierr = panClient.AppMkdirRecursive(familyId, "", "", 0, strings.Split(path.Clean(panDir), "/"))
			if apierr != nil || rs.FileId == "" {
				logger.Verboseln("创建云盘文件夹失败")
				continue
			}
		} else {
			rs = &cloudpan.AppMkdirResult{}
			if familyId > 0 {
				rs.FileId = ""
			} else {
				rs.FileId = "-11"
			}
		}
		dataItem := &dirFileListData{}
		dataItem.Dir = rs

		// files
		param := cloudpan.NewAppFileListParam()
		param.FamilyId = familyId
		param.FileId = rs.FileId
		allFileInfo, err1 := panClient.AppGetAllFileList(param)
		if err1 != nil {
			logger.Verboseln("获取文件信息出错")
			continue
		}
		dataItem.File = allFileInfo

		resultMap[panDir] = dataItem
		time.Sleep(time.Duration(500) * time.Millisecond)
	}
	return resultMap
}