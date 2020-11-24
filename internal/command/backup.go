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

	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"

	"github.com/tickstep/library-go/logger"

	"github.com/tickstep/cloudpan189-api/cloudpan"

	"github.com/tickstep/cloudpan189-go/internal/functions/panupload"

	"github.com/urfave/cli"

	"github.com/tickstep/cloudpan189-go/internal/config"
)

type backupFunc struct {
	User   *config.PanUser
	Client *cloudpan.PanClient
}

func CmdBackup() cli.Command {
	cmd := &backupFunc{}
	return cli.Command{
		Name: "backup",
		Description: `备份指定 <文件/目录> 到云盘 <目标目录> 中

和上传的功能一样，只是备份多进行了如下操作

1. 增加了数据库，记录已经上传的文件信息。
   目前只记录 文件位置、大小、修改时间、MD5 。
2. 上传前先根据数据库记录判断是否需要重新上传。
3. 强制同名覆盖。

注：只备份(上传)新的文件（同名覆盖），不处理删除操作。
`,
		Usage:     "备份文件或目录",
		UsageText: "backup <文件/目录路径1> <文件/目录2> <文件/目录3> ... <目标目录>",
		Category:  "天翼云盘",
		Before:    cmd.Before,
		Action:    cmd.Backup,
		Flags: append(UploadFlags, cli.BoolFlag{
			Name:  "delete",
			Usage: "delete 从网盘中删除本地不存在的文件(默认不删除)",
		}),
	}
}

func (c *backupFunc) Before(cli *cli.Context) error {
	if cli.NArg() < 2 {
		return ErrBadArgs
	}
	User := config.Config.ActiveUser()
	if User == nil {
		return ErrNotLogined
	}
	c.User = User
	c.Client = User.PanClient()
	return nil
}

func OpenSyncDb(path string) (panupload.SyncDb, error) {
	return panupload.OpenSyncDb(config.Config.SyncDBType, path, "ecloud")
}

//根据数据库记录删除本地目录不存在的网盘文件
func (c *backupFunc) DelRemoteFileFromDB(familyId int64, localDir string, savePath string) {
	//先转换为绝对路径，避免出现异常
	localPathDir, err := filepath.Abs(localDir)
	if err == nil {
		localDir = localPathDir
	}

	localPathDir = filepath.Dir(localDir)

	dbpath := filepath.Join(localDir, ".ecloud")

	//数据库目录不存在不需要处理
	if di, err := os.Stat(dbpath); err != nil && !di.IsDir() {
		return
	}

	db, err := OpenSyncDb(dbpath + string(os.PathSeparator) + "db")
	if err != nil {
		fmt.Println("数据库打开失败！", err)
		return
	}
	for ent, err := db.First(savePath); err == nil; ent, err = db.Next(savePath) {
		testPath := strings.TrimPrefix(ent.Path, savePath)
		testPath = filepath.Join(localPathDir, testPath)

		_, err := os.Stat(testPath)
		if err != nil && os.IsNotExist(err) { //本地文件不存在，删除网盘文件
			logger.Verboseln("删除文件", ent.Path)
			var parentId string
			if test := db.Get(path.Dir(ent.Path)); test != nil && test.IsFolder && test.FileID != "" {
				parentId = test.FileID
			}
			if ent.FileID == "" {
				efi, _ := c.Client.AppFileInfoByPath(familyId, ent.Path)
				if efi != nil && efi.FileId != "" {
					ent.FileID = efi.FileId
					parentId = efi.ParentId
				}
			}

			if ent.FileID == "" {
				continue
			}

			if parentId == "" {
				efi, _ := c.Client.AppFileInfoById(familyId, ent.FileID)
				if efi != nil && efi.FileId != "" {
					ent.FileID = efi.FileId
					parentId = efi.ParentId
				}
			}

			var taskId string
			var err *apierror.ApiError

			infoList := cloudpan.BatchTaskInfoList{}
			infoItem := &cloudpan.BatchTaskInfo{
				FileId:      ent.FileID,
				FileName:    path.Base(ent.Path),
				IsFolder:    0,
				SrcParentId: parentId,
			}
			if ent.IsFolder {
				infoItem.IsFolder = 1
			}
			infoList = append(infoList, infoItem)
			delParam := &cloudpan.BatchTaskParam{
				TypeFlag:  cloudpan.BatchTaskTypeDelete,
				TaskInfos: infoList,
			}

			if familyId > 0 {
				taskId, err = c.Client.AppCreateBatchTask(familyId, delParam)
			} else {
				taskId, err = c.Client.CreateBatchTask(delParam)
			}
			if err != nil || taskId == "" {
				fmt.Println("删除网盘文件或目录失败", ent.Path)
			} else {
				db.DelWithPrefix(ent.Path)
				logger.Verboseln("删除网盘文件和数据库记录", testPath, "=>", ent.Path, taskId)
			}
		}
	}
	db.Close()
}

func (c *backupFunc) Backup(cli *cli.Context) error {
	subArgs := cli.Args()
	localpaths := subArgs[:cli.NArg()-1]
	for _, p := range localpaths { //预创建需要的目录
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			os.Mkdir(filepath.Join(p, ".ecloud"), 0644)
		}
	}

	opt := &UploadOptions{
		AllParallel:   cli.Int("p"),
		Parallel:      1, // 天翼云盘一个文件只支持单线程上传
		MaxRetry:      cli.Int("retry"),
		NoRapidUpload: cli.Bool("norapid"),
		NoSplitFile:   true, // 天翼云盘不支持分片并发上传，只支持单线程上传，支持断点续传
		ShowProgress:  !cli.Bool("np"),
		IsOverwrite:   true,
		FamilyId:      parseFamilyId(cli),
	}

	savePath := GetActiveUser().PathJoin(opt.FamilyId, subArgs[cli.NArg()-1])

	RunUpload(localpaths, savePath, opt)
	//根据数据库记录进行同步删除操作
	if cli.Bool("delete") {
		wg := sync.WaitGroup{}
		wg.Add(len(localpaths))
		for _, curPath := range localpaths {
			go func(localdir string) {
				c.DelRemoteFileFromDB(opt.FamilyId, localdir, savePath)
				wg.Done()
			}(curPath)
		}
		wg.Wait()
	}

	return nil
}
