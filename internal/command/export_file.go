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
	"log"
	"os"
	"path"
	"strconv"
	"time"
)

type (
	ImportExportFileItem struct {
		FileMd5 string `json:"md5"`
		FileSize int64 `json:"size"`
		Path string `json:"path"`
		LastOpTime string `json:"lastOpTime"`
	}
)

func RunExportFiles(familyId int64, overwrite bool, panPaths []string, saveLocalFilePath string) {
	activeUser := config.Config.ActiveUser()
	panClient := activeUser.PanClient()

	lfi,_ := os.Stat(saveLocalFilePath)
	realSaveFilePath := saveLocalFilePath
	if lfi != nil {
		if lfi.IsDir() {
			realSaveFilePath = path.Join(saveLocalFilePath, "export_file_") + strconv.FormatInt(time.Now().Unix(), 10) + ".txt"
		} else {
			if !overwrite {
				fmt.Println("导出文件已存在")
				return
			}
		}
	} else {
		// create file
		localDir := path.Dir(saveLocalFilePath)
		dirFs,_ := os.Stat(localDir)
		if dirFs != nil {
			if !dirFs.IsDir() {
				fmt.Println("指定的保存文件路径不合法")
				return
			}
		} else {
			er := os.MkdirAll(localDir, 0755)
			if er != nil {
				fmt.Println("创建文件夹出错")
				return
			}
		}
		realSaveFilePath = saveLocalFilePath
	}

	totalCount := 0
	saveFile, err := os.OpenFile(realSaveFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		log.Fatal(err)
		return
	}

	for _,panPath := range panPaths {
		panPath = activeUser.PathJoin(familyId, panPath)
		panClient.AppFilesDirectoriesRecurseList(familyId, panPath, func(depth int, _ string, fd *cloudpan.AppFileEntity, apiError *apierror.ApiError) bool {
			if apiError != nil {
				logger.Verbosef("%s\n", apiError)
				return true
			}

			// 只需要存储文件即可
			if !fd.IsFolder {
				item := ImportExportFileItem{
					FileMd5: fd.FileMd5,
					FileSize: fd.FileSize,
					Path: fd.Path,
					LastOpTime: fd.LastOpTime,
				}
				jstr,e := json.Marshal(&item)
				if e != nil {
					logger.Verboseln("to json string err")
					return false
				}
				saveFile.WriteString(string(jstr) + "\n")
				totalCount += 1
				time.Sleep(time.Duration(100) * time.Millisecond)
				fmt.Printf("\r导出文件数量: %d", totalCount)
			}
			return true
		})
	}

	// close and save
	if err := saveFile.Close(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\r导出文件总数量: %d\n", totalCount)
	fmt.Printf("导出文件保存路径: %s\n", realSaveFilePath)
}
