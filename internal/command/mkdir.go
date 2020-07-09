package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"path"
	"strings"
)

func RunMkdir(name string) {
	activeUser := GetActiveUser()
	fullpath := activeUser.PathJoin(name)
	pathSlice := strings.Split(fullpath, "/")
	mkSuccess := false
	fileId := ""
	if path.Dir(fullpath) == activeUser.Workdir {
		mkSuccess, fileId = doMkdir(activeUser.WorkdirFileEntity.FileId, path.Clean(path.Dir(fullpath)), len(pathSlice) - 1, pathSlice)
	} else {
		mkSuccess, fileId = doMkdir("", "", 0, pathSlice)
	}

	if mkSuccess {
		fmt.Println("创建文件夹成功: ", fullpath)
	} else {
		if fileId != "" {
			fmt.Println("文件夹已存在: ", fullpath)
		} else {
			fmt.Println("创建文件夹失败: ", fullpath)
		}
	}
}

func doMkdir(parentFileId string, fullPath string, index int, pathSlice []string) (bool, string) {
	activeUser := GetActiveUser()
	if parentFileId == "" {
		// default root "/" entity
		parentFileId = cloudpan.NewFileEntityForRootDir().FileId
		if index == 0 && len(pathSlice) == 1 {
			// root path "/"
			return false, parentFileId
		}

		fullPath = ""
		return doMkdir(parentFileId, fullPath, index + 1, pathSlice)
	}

	if index >= len(pathSlice) {
		return false, parentFileId
	}

	listFilePath := cloudpan.NewFileListParam()
	listFilePath.FileId = parentFileId
	fileResult, err := activeUser.PanClient().FileList(listFilePath)
	if err != nil {
		return false, ""
	}

	// existed?
	for _, fileEntity := range fileResult.Data {
		if fileEntity.FileName == pathSlice[index] {
			return doMkdir(parentFileId, fullPath + "/" + pathSlice[index], index + 1, pathSlice)
		}
	}

	// not existed, mkdir dir
	name := pathSlice[index]
	if !CheckFileNameValid(name) {
		fmt.Println("文件夹名不能包含特殊字符：" + FileNameSpecialChars)
		return false, ""
	}

	r, err := GetActivePanClient().Mkdir(parentFileId, name)
	if err != nil {
		fmt.Printf("创建目录 %s 失败, %s\n", name, err)
		return false, ""
	}

	if (index+1) >= len(pathSlice) {
		return r.IsNew, parentFileId
	} else {
		return doMkdir(r.FileId, fullPath + "/" + pathSlice[index], index + 1, pathSlice)
	}
}
