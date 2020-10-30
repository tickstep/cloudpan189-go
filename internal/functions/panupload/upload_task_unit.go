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
package panupload

import (
	"bytes"
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tickstep/cloudpan189-go/cmder/cmdutil/jsonhelper"

	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/internal/file/uploader"
	"github.com/tickstep/cloudpan189-go/internal/functions"
	"github.com/tickstep/cloudpan189-go/internal/localfile"
	"github.com/tickstep/cloudpan189-go/internal/taskframework"
	"github.com/tickstep/library-go/converter"
	"github.com/tickstep/library-go/requester/rio"
)

type (
	// StepUpload 上传步骤
	StepUpload int

	// UploadTaskUnit 上传的任务单元
	UploadTaskUnit struct {
		LocalFileChecksum *localfile.LocalFileEntity // 要上传的本地文件详情
		Step              StepUpload
		SavePath          string // 保存路径
		FamilyId          int64
		FolderCreateMutex *sync.Mutex
		FolderSyncDb      *FolderSyncDb //文件备份状态数据库

		PanClient         *cloudpan.PanClient
		UploadingDatabase *UploadingDatabase // 数据库
		Parallel          int
		NoRapidUpload     bool // 禁用秒传
		NoSplitFile       bool // 禁用分片上传

		UploadStatistic *UploadStatistic

		taskInfo *taskframework.TaskInfo
		panDir   string
		panFile  string
		state    *uploader.InstanceState

		ShowProgress bool
		IsOverwrite  bool // 覆盖已存在的文件，如果同名文件已存在则移到回收站里
	}
)

const (
	// StepUploadInit 初始化步骤
	StepUploadInit StepUpload = iota
	// 上传前准备，创建上传任务
	StepUploadPrepareUpload
	// StepUploadRapidUpload 秒传步骤
	StepUploadRapidUpload
	// StepUploadUpload 正常上传步骤
	StepUploadUpload
)

const (
	StrUploadFailed = "上传文件失败"
)

func (utu *UploadTaskUnit) SetTaskInfo(taskInfo *taskframework.TaskInfo) {
	utu.taskInfo = taskInfo
}

// prepareFile 解析文件阶段
func (utu *UploadTaskUnit) prepareFile() {
	// 解析文件保存路径
	var (
		panDir, panFile = path.Split(utu.SavePath)
	)
	utu.panDir = path.Clean(panDir)
	utu.panFile = panFile

	// 检测断点续传
	utu.state = utu.UploadingDatabase.Search(&utu.LocalFileChecksum.LocalFileMeta)
	if utu.state != nil || utu.LocalFileChecksum.LocalFileMeta.UploadFileId != "" { // 读取到了上一次上传task请求的fileId
		utu.Step = StepUploadUpload

		// 服务器上次上传的部分数据是否还存在
		var appGetUploadFileStatusResult *cloudpan.AppGetUploadFileStatusResult
		var apierr *apierror.ApiError
		if utu.FamilyId > 0 {
			appGetUploadFileStatusResult, apierr = utu.PanClient.AppFamilyGetUploadFileStatus(utu.FamilyId, utu.LocalFileChecksum.UploadFileId)
		} else {
			appGetUploadFileStatusResult, apierr = utu.PanClient.AppGetUploadFileStatus(utu.LocalFileChecksum.UploadFileId)
		}
		if apierr != nil {
			if apierr.Code == apierror.ApiCodeUploadFileNotFound {
				cmdUploadVerbose.Warn("断点续传失败，需要重新从0开始上传文件：" + apierr.Error())
				utu.Step = StepUploadPrepareUpload
			}
		} else {
			// 需要修正上一次上传值，断点续传
			utu.state.BlockList[0].Range.Begin = appGetUploadFileStatusResult.Size
		}
		return
	}

	if utu.LocalFileChecksum.UploadFileId == "" {
		utu.Step = StepUploadPrepareUpload
		return
	}

	if utu.NoRapidUpload {
		utu.Step = StepUploadUpload
		return
	}

	if utu.LocalFileChecksum.Length > MaxRapidUploadSize {
		fmt.Printf("[%s] 文件超过20GB, 无法使用秒传功能, 跳过秒传...\n", utu.taskInfo.Id())
		utu.Step = StepUploadUpload
		return
	}
	// 下一步: 秒传
	utu.Step = StepUploadRapidUpload
}

// rapidUpload 执行秒传
func (utu *UploadTaskUnit) rapidUpload() (isContinue bool, result *taskframework.TaskUnitRunResult) {
	utu.Step = StepUploadRapidUpload

	// 对于天翼云盘，文件必须是存在才支持秒传
	result = &taskframework.TaskUnitRunResult{}
	fmt.Printf("[%s] 检测秒传中, 请稍候...\n", utu.taskInfo.Id())
	if utu.LocalFileChecksum.FileDataExists == 1 {
		var er *apierror.ApiError
		if utu.FamilyId > 0 {
			_, er = utu.PanClient.AppFamilyUploadFileCommit(utu.FamilyId, utu.LocalFileChecksum.FileCommitUrl, utu.LocalFileChecksum.UploadFileId, utu.LocalFileChecksum.XRequestId)
		} else {
			_, er = utu.PanClient.AppUploadFileCommit(utu.LocalFileChecksum.FileCommitUrl, utu.LocalFileChecksum.UploadFileId, utu.LocalFileChecksum.XRequestId)
		}
		if er != nil {
			result.ResultMessage = "秒传失败"
			result.Err = er
			return true, result
		} else {
			fmt.Printf("[%s] 秒传成功, 保存到网盘路径: %s\n\n", utu.taskInfo.Id(), utu.SavePath)
			result.Succeed = true
			return false, result
		}
	} else {
		fmt.Printf("[%s] 秒传失败，开始正常上传文件\n", utu.taskInfo.Id())
		result.Succeed = false
		result.ResultMessage = "文件未曾上传，无法秒传"
		return true, result
	}
}

// upload 上传文件
func (utu *UploadTaskUnit) upload() (result *taskframework.TaskUnitRunResult) {
	utu.Step = StepUploadUpload

	var blockSize int64
	if utu.NoSplitFile {
		// 不分片上传，天翼网盘不支持分片，所以正常应该到这个分支
		blockSize = utu.LocalFileChecksum.Length
	} else {
		blockSize = getBlockSize(utu.LocalFileChecksum.Length)
	}

	muer := uploader.NewMultiUploader(utu.LocalFileChecksum.FileUploadUrl, utu.LocalFileChecksum.FileCommitUrl, utu.LocalFileChecksum.UploadFileId, utu.LocalFileChecksum.XRequestId,
		NewPanUpload(utu.PanClient, utu.SavePath, utu.LocalFileChecksum.FileUploadUrl, utu.LocalFileChecksum.FileCommitUrl, utu.LocalFileChecksum.UploadFileId, utu.LocalFileChecksum.XRequestId, utu.FamilyId),
		rio.NewFileReaderAtLen64(utu.LocalFileChecksum.GetFile()), &uploader.MultiUploaderConfig{
			Parallel:  utu.Parallel,
			BlockSize: blockSize,
			MaxRate:   config.Config.MaxUploadRate,
		})

	// 设置断点续传
	if utu.state != nil {
		muer.SetInstanceState(utu.state)
	}

	muer.OnUploadStatusEvent(func(status uploader.Status, updateChan <-chan struct{}) {
		select {
		case <-updateChan:
			utu.UploadingDatabase.UpdateUploading(&utu.LocalFileChecksum.LocalFileMeta, muer.InstanceState())
			utu.UploadingDatabase.Save()
		default:
		}

		if utu.ShowProgress {
			fmt.Printf("\r[%s] ↑ %s/%s %s/s in %s ............", utu.taskInfo.Id(),
				converter.ConvertFileSize(status.Uploaded(), 2),
				converter.ConvertFileSize(status.TotalSize(), 2),
				converter.ConvertFileSize(status.SpeedsPerSecond(), 2),
				status.TimeElapsed(),
			)
		}
	})

	// result
	result = &taskframework.TaskUnitRunResult{}
	muer.OnSuccess(func() {
		fmt.Printf("\n")
		fmt.Printf("[%s] 上传文件成功, 保存到网盘路径: %s\n", utu.taskInfo.Id(), utu.SavePath)
		// 统计
		utu.UploadStatistic.AddTotalSize(utu.LocalFileChecksum.Length)
		utu.UploadingDatabase.Delete(&utu.LocalFileChecksum.LocalFileMeta) // 删除
		utu.UploadingDatabase.Save()
		result.Succeed = true
	})
	muer.OnError(func(err error) {
		apiError, ok := err.(*apierror.ApiError)
		if !ok {
			// 未知错误类型 (非预期的)
			// 不重试
			result.ResultMessage = "上传文件错误"
			result.Err = err
			return
		}

		// 默认需要重试
		result.NeedRetry = true

		switch apiError.ErrCode() {
		default:
			result.ResultMessage = StrUploadFailed
			result.NeedRetry = false
			result.Err = apiError
		}
		return
	})
	muer.Execute()

	return
}

func (utu *UploadTaskUnit) OnRetry(lastRunResult *taskframework.TaskUnitRunResult) {
	// 输出错误信息
	if lastRunResult.Err == nil {
		// result中不包含Err, 忽略输出
		fmt.Printf("[%s] %s, 重试 %d/%d\n", utu.taskInfo.Id(), lastRunResult.ResultMessage, utu.taskInfo.Retry(), utu.taskInfo.MaxRetry())
		return
	}
	fmt.Printf("[%s] %s, %s, 重试 %d/%d\n", utu.taskInfo.Id(), lastRunResult.ResultMessage, lastRunResult.Err, utu.taskInfo.Retry(), utu.taskInfo.MaxRetry())
}

func (utu *UploadTaskUnit) OnSuccess(lastRunResult *taskframework.TaskUnitRunResult) {
}

func (utu *UploadTaskUnit) OnFailed(lastRunResult *taskframework.TaskUnitRunResult) {
	// 失败
}

var ResultLocalFileNotUpdated = &taskframework.TaskUnitRunResult{ResultCode: 1, Succeed: true, ResultMessage: "本地文件未更新，无需上传！"}
var ResultUpdateLocalDatabase = &taskframework.TaskUnitRunResult{ResultCode: 2, Succeed: true, ResultMessage: "本地文件和云端文件MD5一致，无需上传！"}

func (utu *UploadTaskUnit) OnComplete(lastRunResult *taskframework.TaskUnitRunResult) {
	if utu.FolderSyncDb == nil || !lastRunResult.Succeed || lastRunResult == ResultLocalFileNotUpdated { //不需要更新数据库
		return
	}
	var buf bytes.Buffer
	jsonhelper.MarshalData(&buf, localfile.LocalFileMeta{
		MD5:     utu.LocalFileChecksum.MD5,
		ModTime: utu.LocalFileChecksum.ModTime,
		Length:  utu.LocalFileChecksum.Length,
	})
	utu.FolderSyncDb.Put(utu.SavePath, buf.Bytes())
}

func (utu *UploadTaskUnit) RetryWait() time.Duration {
	return functions.RetryWait(utu.taskInfo.Retry())
}

func (utu *UploadTaskUnit) Run() (result *taskframework.TaskUnitRunResult) {

	err := utu.LocalFileChecksum.OpenPath()
	if err != nil {
		fmt.Printf("[%s] 文件不可读, 错误信息: %s, 跳过...\n", utu.taskInfo.Id(), err)
		return
	}
	defer utu.LocalFileChecksum.Close() // 关闭文件

	timeStart := time.Now()
	result = &taskframework.TaskUnitRunResult{}

	fmt.Printf("%s [%s] 准备上传: %s\n", time.Now().Format("2006-01-02 15:04:06"), utu.taskInfo.Id(), utu.LocalFileChecksum.Path)

	defer func() {
		var msg string
		if result.Err != nil {
			msg = "失败！" + result.ResultMessage + "," + result.Err.Error()
		} else if result.Succeed {
			msg = "成功！" + result.ResultMessage
		} else {
			msg = result.ResultMessage
		}
		fmt.Printf("%s [%s] 文件上传结果：%s  耗时 %s\n", time.Now().Format("2006-01-02 15:04:06"), utu.taskInfo.Id(), msg, time.Now().Sub(timeStart))
	}()
	// 准备文件
	utu.prepareFile()

	var r *cloudpan.AppCreateUploadFileResult
	var apierr *apierror.ApiError
	var rs *cloudpan.AppMkdirResult
	var appCreateUploadFileParam *cloudpan.AppCreateUploadFileParam
	var md5Str string
	var saveFilePath string
	var testFileMeta *localfile.LocalFileMeta = &localfile.LocalFileMeta{}

	switch utu.Step {
	case StepUploadPrepareUpload:
		goto StepUploadPrepareUpload
	case StepUploadRapidUpload:
		goto stepUploadRapidUpload
	case StepUploadUpload:
		goto stepUploadUpload
	}

StepUploadPrepareUpload:

	if utu.FolderSyncDb != nil {
		//启用了备份功能，强制使用覆盖同名文件功能
		utu.IsOverwrite = true
		data := utu.FolderSyncDb.Get(utu.SavePath)
		if data != nil {
			jsonhelper.UnmarshalData(bytes.NewBuffer(data), testFileMeta)
		}
	}
	//文件是否有更新判断，这边只简单判断文件日期和文件大小
	if testFileMeta.ModTime == utu.LocalFileChecksum.ModTime &&
		testFileMeta.Length == utu.LocalFileChecksum.Length {
		return ResultLocalFileNotUpdated
	}
	// 创建上传任务
	utu.LocalFileChecksum.Sum(localfile.CHECKSUM_MD5)

	if testFileMeta.MD5 == utu.LocalFileChecksum.MD5 {
		return ResultUpdateLocalDatabase
	}

	utu.FolderCreateMutex.Lock()
	saveFilePath = path.Dir(utu.SavePath)
	if saveFilePath != "/" {
		rs, apierr = utu.PanClient.AppMkdirRecursive(utu.FamilyId, "", "", 0, strings.Split(path.Clean(saveFilePath), "/"))
		if apierr != nil || rs.FileId == "" {
			result.Err = apierr
			result.ResultMessage = "创建云盘文件夹失败"
			return
		}
	} else {
		rs = &cloudpan.AppMkdirResult{}
		if utu.FamilyId > 0 {
			rs.FileId = ""
		} else {
			rs.FileId = "-11"
		}
	}
	time.Sleep(time.Duration(2) * time.Second)
	utu.FolderCreateMutex.Unlock()

	if utu.IsOverwrite {
		// 标记覆盖旧同名文件
		// 检查同名文件是否存在
		efi, apierr := utu.PanClient.AppFileInfoByPath(utu.FamilyId, utu.SavePath)
		if apierr != nil && apierr.Code != apierror.ApiCodeFileNotFoundCode {
			result.Err = apierr
			result.ResultMessage = "检测同名文件失败"
			return
		}
		if efi != nil && efi.FileId != "" {
			if efi.FileMd5 == strings.ToUpper(utu.LocalFileChecksum.MD5) {
				return ResultUpdateLocalDatabase
			}
			// existed, delete it
			infoList := cloudpan.BatchTaskInfoList{}
			isFolder := 0
			if efi.IsFolder {
				isFolder = 1
			}
			infoItem := &cloudpan.BatchTaskInfo{
				FileId:      efi.FileId,
				FileName:    efi.FileName,
				IsFolder:    isFolder,
				SrcParentId: efi.ParentId,
			}
			infoList = append(infoList, infoItem)
			delParam := &cloudpan.BatchTaskParam{
				TypeFlag:  cloudpan.BatchTaskTypeDelete,
				TaskInfos: infoList,
			}

			var taskId string
			var err *apierror.ApiError
			if utu.FamilyId > 0 {
				taskId, err = utu.PanClient.AppCreateBatchTask(utu.FamilyId, delParam)
			} else {
				taskId, err = utu.PanClient.CreateBatchTask(delParam)
			}

			if err != nil || taskId == "" {
				result.Err = err
				result.ResultMessage = "无法删除文件，请稍后重试"
				return
			}
			time.Sleep(time.Duration(500) * time.Millisecond)
			fmt.Println(utu.taskInfo.Id(), "检测到同名文件，已移动到回收站: "+utu.SavePath)
		}
	}

	md5Str = utu.LocalFileChecksum.MD5
	if utu.LocalFileChecksum.Length == 0 {
		md5Str = cloudpan.DefaultEmptyFileMd5
	}

	appCreateUploadFileParam = &cloudpan.AppCreateUploadFileParam{
		ParentFolderId: rs.FileId,
		FileName:       filepath.Base(utu.LocalFileChecksum.Path),
		Size:           utu.LocalFileChecksum.Length,
		Md5:            md5Str,
		LastWrite:      time.Unix(utu.LocalFileChecksum.ModTime, 0).Format("2006-01-02 15:04:05"),
		LocalPath:      utu.LocalFileChecksum.Path,
		FamilyId:       utu.FamilyId,
	}
	if utu.FamilyId > 0 {
		r, apierr = utu.PanClient.AppFamilyCreateUploadFile(appCreateUploadFileParam)
	} else {
		r, apierr = utu.PanClient.AppCreateUploadFile(appCreateUploadFileParam)
	}
	if apierr != nil {
		result.Err = apierr
		result.ResultMessage = "创建上传任务失败"
		return
	}

	utu.LocalFileChecksum.ParentFolderId = rs.FileId
	utu.LocalFileChecksum.UploadFileId = r.UploadFileId
	utu.LocalFileChecksum.FileUploadUrl = r.FileUploadUrl
	utu.LocalFileChecksum.FileCommitUrl = r.FileCommitUrl
	utu.LocalFileChecksum.FileDataExists = r.FileDataExists
	utu.LocalFileChecksum.XRequestId = r.XRequestId

stepUploadRapidUpload:
	// 秒传
	if !utu.NoRapidUpload {
		isContinue, rapidUploadResult := utu.rapidUpload()
		if !isContinue {
			// 秒传成功, 返回秒传的结果
			return rapidUploadResult
		}
	}

stepUploadUpload:
	// 正常上传流程
	uploadResult := utu.upload()

	return uploadResult
}
