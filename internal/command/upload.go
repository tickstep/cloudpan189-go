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
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/tickstep/library-go/logger"

	"github.com/urfave/cli"

	"github.com/tickstep/cloudpan189-go/cmder/cmdutil"

	"github.com/oleiade/lane"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/internal/functions/panupload"
	"github.com/tickstep/cloudpan189-go/internal/localfile"
	"github.com/tickstep/cloudpan189-go/internal/taskframework"
	"github.com/tickstep/library-go/converter"
)

const (
	// DefaultUploadMaxAllParallel 默认所有文件并发上传数量，即可以同时并发上传多少个文件
	DefaultUploadMaxAllParallel = 1
	// DefaultUploadMaxRetry 默认上传失败最大重试次数
	DefaultUploadMaxRetry = 3
)

type (
	// UploadOptions 上传可选项
	UploadOptions struct {
		AllParallel   int // 所有文件并发上传数量，即可以同时并发上传多少个文件
		Parallel      int // 单个文件并发上传数量
		MaxRetry      int
		NoRapidUpload bool
		NoSplitFile   bool // 禁用分片上传
		ShowProgress  bool
		IsOverwrite   bool // 覆盖已存在的文件，如果同名文件已存在则移到回收站里
		FamilyId      int64
	}
)

var UploadFlags = []cli.Flag{
	cli.IntFlag{
		Name:  "p",
		Usage: "本次操作文件上传并发数量，即可以同时并发上传多少个文件。0代表跟从配置文件设置",
		Value: 0,
	},
	cli.IntFlag{
		Name:  "retry",
		Usage: "上传失败最大重试次数",
		Value: DefaultUploadMaxRetry,
	},
	cli.BoolFlag{
		Name:  "np",
		Usage: "no progress 不展示下载进度条",
	},
	cli.BoolFlag{
		Name:  "ow",
		Usage: "overwrite, 覆盖已存在的同名文件，注意已存在的文件会被移到回收站",
	},
	cli.BoolFlag{
		Name:  "norapid",
		Usage: "不检测秒传",
	},
	cli.StringFlag{
		Name:  "familyId",
		Usage: "家庭云ID",
		Value: "",
	},
}

// RunUpload 执行文件上传
func RunUpload(localPaths []string, savePath string, opt *UploadOptions) {
	activeUser := GetActiveUser()
	if opt == nil {
		opt = &UploadOptions{}
	}

	// 检测opt
	if opt.AllParallel <= 0 {
		opt.AllParallel = config.Config.MaxUploadParallel
	}
	if opt.Parallel <= 0 {
		opt.Parallel = 1
	}

	if opt.MaxRetry < 0 {
		opt.MaxRetry = DefaultUploadMaxRetry
	}

	savePath = activeUser.PathJoin(opt.FamilyId, savePath)
	_, err1 := activeUser.PanClient().AppFileInfoByPath(opt.FamilyId, savePath)
	if err1 != nil {
		fmt.Printf("警告: 上传文件, 获取云盘路径 %s 错误, %s\n", savePath, err1)
	}

	switch len(localPaths) {
	case 0:
		fmt.Printf("本地路径为空\n")
		return
	}

	// 打开上传状态
	uploadDatabase, err := panupload.NewUploadingDatabase()
	if err != nil {
		fmt.Printf("打开上传未完成数据库错误: %s\n", err)
		return
	}
	defer uploadDatabase.Close()

	var (
		// 使用 task framework
		executor = &taskframework.TaskExecutor{
			IsFailedDeque: true, // 失败统计
		}
		// 统计
		statistic = &panupload.UploadStatistic{}

		folderCreateMutex = &sync.Mutex{}
	)
	executor.SetParallel(opt.AllParallel)

	statistic.StartTimer() // 开始计时

	wg := sync.WaitGroup{}

	// 启动上传任务
	Done := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		var failedList []*lane.Deque
	FOR:
		for {
			select {
			case <-Done:
				break FOR
			default:
				executor.Execute()
				failed := executor.FailedDeque()
				if failed.Size() > 0 {
					failedList = append(failedList, failed)
				}
			}
		}
		fmt.Printf("\n")
		fmt.Printf("上传结束, 时间: %s, 总大小: %s\n", statistic.Elapsed()/1e6*1e6, converter.ConvertFileSize(statistic.TotalSize()))

		// 输出上传失败的文件列表
		for _, failed := range failedList {
			if failed.Size() != 0 {
				fmt.Printf("以下文件上传失败: \n")
				tb := cmdtable.NewTable(os.Stdout)
				for e := failed.Shift(); e != nil; e = failed.Shift() {
					item := e.(*taskframework.TaskInfoItem)
					tb.Append([]string{item.Info.Id(), item.Unit.(*panupload.UploadTaskUnit).LocalFileChecksum.Path})
				}
				tb.Render()
			}
		}
	}()

	for _, curPath := range localPaths {
		var walkFunc filepath.WalkFunc
		var db *panupload.FolderSyncDb
		curPath = filepath.Clean(curPath)
		localPathDir := filepath.Dir(curPath)

		// 避免去除文件名开头的"."
		if localPathDir == "." {
			localPathDir = ""
		}

		if fi, err := os.Stat(curPath); err == nil && fi.IsDir() {
			//使用绝对路径避免异常
			dbpath, err := filepath.Abs(curPath)
			if err != nil {
				dbpath = curPath
			}
			dbpath += string(os.PathSeparator) + ".ecloud"
			if di, err := os.Stat(dbpath); err == nil && di.IsDir() {
				db, err = panupload.OpenSyncDb(dbpath+string(os.PathSeparator)+"db", "ecloud")
				if db != nil {
					defer func(syncDb *panupload.FolderSyncDb) {
						db.Close()
					}(db)
				} else {
					fmt.Println(err)
				}
			}
		}

		walkFunc = func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if fi.IsDir() { // 忽略目录
				if fi.Name() == ".ecloud" {
					return filepath.SkipDir
				}
				return nil
			}

			if fi.Mode()&os.ModeSymlink != 0 { // 读取 symbol link
				err = filepath.Walk(file+string(os.PathSeparator), walkFunc)
				return err
			}

			subSavePath := strings.TrimPrefix(file, localPathDir)

			// 针对 windows 的目录处理
			if os.PathSeparator == '\\' {
				subSavePath = cmdutil.ConvertToUnixPathSeparator(subSavePath)
			}

			subSavePath = path.Clean(savePath + cloudpan.PathSeparator + subSavePath)

			if db != nil {
				if test := db.Get(subSavePath); test != nil && test.Size == fi.Size() && test.ModTime == fi.ModTime().Unix() {
					logger.Verbosef("文件未修改跳过:%s\n", file)
					return nil
				}
			}

			taskinfo := executor.Append(&panupload.UploadTaskUnit{
				LocalFileChecksum: localfile.NewLocalFileEntity(file),
				SavePath:          subSavePath,
				FamilyId:          opt.FamilyId,
				PanClient:         activeUser.PanClient(),
				UploadingDatabase: uploadDatabase,
				FolderCreateMutex: folderCreateMutex,
				Parallel:          opt.Parallel,
				NoRapidUpload:     opt.NoRapidUpload,
				NoSplitFile:       opt.NoSplitFile,
				UploadStatistic:   statistic,
				ShowProgress:      opt.ShowProgress,
				IsOverwrite:       opt.IsOverwrite,
				FolderSyncDb:      db,
			}, opt.MaxRetry)

			fmt.Printf("%s [%s] 加入上传队列: %s\n", time.Now().Format("2006-01-02 15:04:05"), taskinfo.Id(), file)
			return nil
		}
		if err := filepath.Walk(curPath, walkFunc); err != nil {
			fmt.Printf("警告: 遍历错误: %s\n", err)
		}
	}
	close(Done)
	wg.Wait()
	logger.Verboseln("Upload Ok!")
}

func RunRapidUpload(familyId int64, isOverwrite bool, panFilePath string, md5Str string, length int64) {
	activeUser := GetActiveUser()
	panClient := activeUser.PanClient()

	if length == 0 {
		md5Str = cloudpan.DefaultEmptyFileMd5
	}

	var r *cloudpan.AppCreateUploadFileResult
	var apierr *apierror.ApiError
	var rs *cloudpan.AppMkdirResult
	var appCreateUploadFileParam *cloudpan.AppCreateUploadFileParam
	var saveFilePath string

	saveFilePath = activeUser.PathJoin(familyId, panFilePath)
	panDir, panFileName := path.Split(saveFilePath)
	if panDir != "/" {
		rs, apierr = panClient.AppMkdirRecursive(familyId, "", "", 0, strings.Split(path.Clean(panDir), "/"))
		if apierr != nil || rs.FileId == "" {
			fmt.Println("创建云盘文件夹失败")
			return
		}
	} else {
		rs = &cloudpan.AppMkdirResult{}
		if familyId > 0 {
			rs.FileId = ""
		} else {
			rs.FileId = "-11"
		}
	}
	time.Sleep(time.Duration(2) * time.Second)

	if isOverwrite {
		// 标记覆盖旧同名文件
		// 检查同名文件是否存在
		efi, apierr := panClient.AppFileInfoByPath(familyId, saveFilePath)
		if apierr != nil && apierr.Code != apierror.ApiCodeFileNotFoundCode {
			fmt.Println("检测同名文件失败，请稍后重试")
			return
		}
		if efi != nil && efi.FileId != "" {
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
			if familyId > 0 {
				taskId, err = panClient.AppCreateBatchTask(familyId, delParam)
			} else {
				taskId, err = panClient.CreateBatchTask(delParam)
			}

			if err != nil || taskId == "" {
				fmt.Println("无法删除文件，请稍后重试")
				return
			}
			time.Sleep(time.Duration(500) * time.Millisecond)
			fmt.Println("检测到同名文件，已移动到回收站: " + saveFilePath)
		}
	}

	appCreateUploadFileParam = &cloudpan.AppCreateUploadFileParam{
		ParentFolderId: rs.FileId,
		FileName:       panFileName,
		Size:           length,
		Md5:            strings.ToUpper(md5Str),
		LastWrite:      time.Now().Format("2006-01-02 15:04:05"),
		LocalPath:      "",
		FamilyId:       familyId,
	}
	if familyId > 0 {
		r, apierr = panClient.AppFamilyCreateUploadFile(appCreateUploadFileParam)
	} else {
		r, apierr = panClient.AppCreateUploadFile(appCreateUploadFileParam)
	}
	if apierr != nil {
		fmt.Println("创建上传任务失败")
		return
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
			return
		} else {
			fmt.Printf("秒传成功, 保存到网盘路径: %s\n", saveFilePath)
		}
	} else {
		fmt.Printf("文件未曾上传，无法秒传\n")
	}
}
