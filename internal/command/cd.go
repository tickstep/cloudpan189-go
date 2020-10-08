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
	"github.com/tickstep/cloudpan189-go/internal/config"
)

func RunChangeDirectory(familyId int64, targetPath string) {
	user := config.Config.ActiveUser()
	targetPath = user.PathJoin(familyId, targetPath)

	targetPathInfo, err := user.PanClient().AppFileInfoByPath(familyId, targetPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !targetPathInfo.IsFolder {
		fmt.Printf("错误: %s 不是一个目录 (文件夹)\n", targetPath)
		return
	}

	if IsFamilyCloud(familyId) {
		user.FamilyWorkdir = targetPath
		user.FamilyWorkdirFileEntity = *targetPathInfo
	} else {
		user.Workdir = targetPath
		user.WorkdirFileEntity = *targetPathInfo
	}

	fmt.Printf("改变工作目录: %s\n", targetPath)
}
