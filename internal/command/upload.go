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
	"github.com/tickstep/cloudpan189-go/cmder"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
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
		ExcludeNames []string // 排除的文件名，包括文件夹和文件。即这些文件/文件夹不进行上传，支持正则表达式
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
		Usage: "no progress 不展示上传进度条",
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
	cli.StringSliceFlag{
		Name:  "exn",
		Usage: "exclude name，指定排除的文件夹或者文件的名称，只支持正则表达式。支持同时排除多个名称，每一个名称就是一个exn参数",
		Value: nil,
	},
}

func CmdUpload() cli.Command {
	return cli.Command{
		Name:      "upload",
		Aliases:   []string{"u"},
		Usage:     "上传文件/目录",
		UsageText: cmder.App().Name + " upload <本地文件/目录的路径1> <文件/目录2> <文件/目录3> ... <目标目录>",
		Description: `
	上传指定的文件夹或者文件，上传的文件将会保存到 <目标目录>.

  示例:
    1. 将本地的 C:\Users\Administrator\Desktop\1.mp4 上传到网盘 /视频 目录
    注意区别反斜杠 "\" 和 斜杠 "/" !!!
    cloudpan189-go upload C:/Users/Administrator/Desktop/1.mp4 /视频

    2. 将本地的 C:\Users\Administrator\Desktop\1.mp4 和 C:\Users\Administrator\Desktop\2.mp4 上传到网盘 /视频 目录
    cloudpan189-go upload C:/Users/Administrator/Desktop/1.mp4 C:/Users/Administrator/Desktop/2.mp4 /视频

    3. 将本地的 C:\Users\Administrator\Desktop 整个目录上传到网盘 /视频 目录
    cloudpan189-go upload C:/Users/Administrator/Desktop /视频

    4. 使用相对路径
    cloudpan189-go upload 1.mp4 /视频

    5. 覆盖上传，已存在的同名文件会被移到回收站
    cloudpan189-go upload -ow 1.mp4 /视频

    6. 将本地的 C:\Users\Administrator\Video 整个目录上传到网盘 /视频 目录，但是排除所有的.jpg文件
    cloudpan189-go upload -exn "\.jpg$" C:/Users/Administrator/Video /视频

    7. 将本地的 C:\Users\Administrator\Video 整个目录上传到网盘 /视频 目录，但是排除所有的.jpg文件和.mp3文件，每一个排除项就是一个exn参数
    cloudpan189-go upload -exn "\.jpg$" -exn "\.mp3$" C:/Users/Administrator/Video /视频

    8. 将本地的 C:\Users\Administrator\Video 整个目录上传到网盘 /视频 目录，但是排除所有的 @eadir 文件夹
    cloudpan189-go upload -exn "^@eadir$" C:/Users/Administrator/Video /视频

  参考：
    以下是典型的排除特定文件或者文件夹的例子，注意：参数值必须是正则表达式。在正则表达式中，^表示匹配开头，$表示匹配结尾。
    1)排除@eadir文件或者文件夹：-exn "^@eadir$"
    2)排除.jpg文件：-exn "\.jpg$"
    3)排除.号开头的文件：-exn "^\."
    4)排除 myfile.txt 文件：-exn "^myfile.txt$"
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NArg() < 2 {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}

			subArgs := c.Args()
			RunUpload(subArgs[:c.NArg()-1], subArgs[c.NArg()-1], &UploadOptions{
				AllParallel:   c.Int("p"),
				Parallel:      1, // 天翼云盘一个文件只支持单线程上传
				MaxRetry:      c.Int("retry"),
				NoRapidUpload: c.Bool("norapid"),
				NoSplitFile:   true, // 天翼云盘不支持分片并发上传，只支持单线程上传，支持断点续传
				ShowProgress:  !c.Bool("np"),
				IsOverwrite:   c.Bool("ow"),
				FamilyId:      parseFamilyId(c),
				ExcludeNames: c.StringSlice("exn"),
			})
			return nil
		},
		Flags: UploadFlags,
	}
}

func CmdRapidUpload() cli.Command {
	return cli.Command{
		Name:      "rapidupload",
		Aliases:   []string{"ru"},
		Usage:     "手动秒传文件",
		UsageText: cmder.App().Name + " rapidupload -size=<文件的大小> -md5=<文件的md5值> <保存的网盘路径, 需包含文件名>",
		Description: `
	使用此功能秒传文件, 前提是知道文件的大小, md5, 且网盘中存在一模一样的文件.
	上传的文件将会保存到网盘的目标目录.

	示例:

	1. 如果秒传成功, 则保存到网盘路径 /test/file.txt
	cloudpan189-go rapidupload -size=56276137 -md5=fbe082d80e90f90f0fb1f94adbbcfa7f /test/file.txt
`,
		Category: "天翼云盘",
		Before:   cmder.ReloadConfigFunc,
		Action: func(c *cli.Context) error {
			if c.NArg() <= 0 || !c.IsSet("md5") || !c.IsSet("size") {
				cli.ShowCommandHelp(c, c.Command.Name)
				return nil
			}

			RunRapidUpload(parseFamilyId(c), c.Bool("ow"), c.Args().Get(0), c.String("md5"), c.Int64("size"))
			return nil
		},
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:     "md5",
				Usage:    "文件的 md5 值",
				Required: true,
			},
			cli.Int64Flag{
				Name:     "size",
				Usage:    "文件的大小",
				Required: true,
			},
			cli.BoolFlag{
				Name:  "ow",
				Usage: "overwrite, 覆盖已存在的文件",
			},
			cli.StringFlag{
				Name:  "familyId",
				Usage: "家庭云ID",
				Value: "",
			},
		},
	}
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

	// 遍历指定的文件并创建上传任务
	for _, curPath := range localPaths {
		var walkFunc filepath.WalkFunc
		var db panupload.SyncDb
		curPath = filepath.Clean(curPath)
		localPathDir := filepath.Dir(curPath)

		// 是否排除上传
		if isExcludeFile(curPath, opt) {
			fmt.Printf("排除文件: %s\n", curPath)
			continue
		}

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
					defer func(syncDb panupload.SyncDb) {
						db.Close()
					}(db)
				} else {
					fmt.Println(curPath, "同步数据库打开失败,跳过该目录的备份", err)
					continue
				}
			}
		}

		walkFunc = func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 是否排除上传
			if isExcludeFile(file, opt) {
				fmt.Printf("排除文件: %s\n", file)
				return filepath.SkipDir
			}

			if fi.Mode()&os.ModeSymlink != 0 { // 读取 symbol link
				err = WalkAllFile(file+string(os.PathSeparator), walkFunc)
				return err
			}

			subSavePath := strings.TrimPrefix(file, localPathDir)

			// 针对 windows 的目录处理
			if os.PathSeparator == '\\' {
				subSavePath = cmdutil.ConvertToUnixPathSeparator(subSavePath)
			}

			subSavePath = path.Clean(savePath + cloudpan.PathSeparator + subSavePath)
			var ufm *panupload.UploadedFileMeta

			if db != nil {
				if ufm = db.Get(subSavePath); ufm.Size == fi.Size() && ufm.ModTime == fi.ModTime().Unix() {
					logger.Verbosef("文件未修改跳过:%s\n", file)
					return nil
				}
			}

			if fi.IsDir() { // 备份目录处理
				if strings.HasPrefix(fi.Name(), ".ecloud") {
					return filepath.SkipDir
				}
				//不存在同步数据库时跳过
				if db == nil || ufm.FileID != "" {
					return nil
				}
				panClient := activeUser.PanClient()
				fmt.Println(subSavePath, "云盘文件夹预创建")
				//首先尝试直接创建文件夹
				if ufm = db.Get(path.Dir(subSavePath)); ufm.IsFolder == true && ufm.FileID != "" {
					rs, err := panClient.AppMkdir(opt.FamilyId, ufm.FileID, fi.Name())
					if err == nil && rs != nil && rs.FileId != "" {
						db.Put(subSavePath, &panupload.UploadedFileMeta{FileID: rs.FileId, IsFolder: true, ModTime: fi.ModTime().Unix(), Rev: rs.Rev, ParentId: rs.ParentId})
						return nil
					}
				}
				rs, err := panClient.AppMkdirRecursive(opt.FamilyId, "", "", 0, strings.Split(path.Clean(subSavePath), "/"))
				if err == nil && rs != nil && rs.FileId != "" {
					db.Put(subSavePath, &panupload.UploadedFileMeta{FileID: rs.FileId, IsFolder: true, ModTime: fi.ModTime().Unix(), Rev: rs.Rev, ParentId: rs.ParentId})
					return nil
				}
				fmt.Println(subSavePath, "创建云盘文件夹失败", err)
				return filepath.SkipDir
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
		if err := WalkAllFile(curPath, walkFunc); err != nil {
			fmt.Printf("警告: 遍历错误: %s\n", err)
		}
	}
	close(Done)
	wg.Wait()
	logger.Verboseln("Upload Ok!")
}

// 是否是排除上传的文件
func isExcludeFile(filePath string, opt *UploadOptions) bool {
	if opt == nil || len(opt.ExcludeNames) == 0{
		return false
	}

	for _,pattern := range opt.ExcludeNames {
		fileName := path.Base(filePath)

		m,_ := regexp.MatchString(pattern, fileName)
		if m {
			return true
		}
	}
	return false
}

func WalkAllFile(dirPath string, walkFn filepath.WalkFunc) error {
	info, err := os.Lstat(dirPath)
	if err != nil {
		err = walkFn(dirPath, nil, err)
	} else {
		err = walkAllFile(dirPath, info, walkFn)
	}
	return err
}

func walkAllFile(dirPath string, info os.FileInfo, walkFn filepath.WalkFunc) error {
	if !info.IsDir() {
		return walkFn(dirPath, info, nil)
	}

	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return walkFn(dirPath, nil, err)
	}
	for _, fi := range files {
		subFilePath := dirPath + "/" + fi.Name()
		err := walkFn(subFilePath, fi, err)
		if err != nil && err != filepath.SkipDir {
			return err
		}
		if fi.IsDir() {
			if err == filepath.SkipDir {
				continue
			}
			err := walkAllFile(subFilePath, fi, walkFn)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
		fmt.Println("创建上传任务失败：", apierr.Error())
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
