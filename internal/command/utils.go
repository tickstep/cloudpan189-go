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
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/library-go/logger"
	"path"
)

var (
	panCommandVerbose = logger.New("PANCOMMAND", config.EnvVerbose)
)

// GetFileInfoByPaths 获取指定文件路径的文件详情信息
func GetAppFileInfoByPaths(familyId int64, paths ...string) (fileInfoList []*cloudpan.AppFileEntity, failedPaths []string, error error) {
	if len(paths) <= 0 {
		return nil, nil, fmt.Errorf("请指定文件路径")
	}
	activeUser := GetActiveUser()

	for idx := 0; idx < len(paths); idx++ {
		absolutePath := path.Clean(activeUser.PathJoin(familyId, paths[idx]))
		fe, err := activeUser.PanClient().AppFileInfoByPath(familyId, absolutePath)
		if err != nil {
			failedPaths = append(failedPaths, absolutePath)
			continue
		}
		fileInfoList = append(fileInfoList, fe)
	}
	return
}

// GetFileInfoByPaths 获取指定文件路径的文件详情信息
func GetFileInfoByPaths(paths ...string) (fileInfoList []*cloudpan.FileEntity, failedPaths []string, error error) {
	if len(paths) <= 0 {
		return nil, nil, fmt.Errorf("请指定文件路径")
	}
	activeUser := GetActiveUser()

	for idx := 0; idx < len(paths); idx++ {
		absolutePath := path.Clean(activeUser.PathJoin(0, paths[idx]))
		fe, err := activeUser.PanClient().FileInfoByPath(absolutePath)
		if err != nil {
			failedPaths = append(failedPaths, absolutePath)
			continue
		}
		fileInfoList = append(fileInfoList, fe)
	}
	return
}

// matchPathByShellPattern 通配符匹配路径，允许返回多个匹配结果
func matchPathByShellPattern(familyId int64, patterns ...string) (files []*cloudpan.AppFileEntity, e error) {
	acUser := GetActiveUser()
	for k := range patterns {
		ps, err := acUser.PanClient().MatchPathByShellPattern(familyId, acUser.PathJoin(familyId, patterns[k]))
		if err != nil {
			return nil, err
		}
		files = append(files, *ps...)
	}
	return files, nil
}

func makePathAbsolute(familyId int64, patterns ...string) (panpaths []string, err error) {
	acUser := GetActiveUser()
	for k := range patterns {
		ps := acUser.PathJoin(familyId, patterns[k])
		panpaths = append(panpaths, ps)
	}
	return panpaths, nil
}

func IsFamilyCloud(familyId int64) bool {
	return familyId > 0
}

func GetFamilyCloudMark(familyId int64) string {
	if familyId > 0 {
		return "家庭云"
	}
	return "个人云"
}
