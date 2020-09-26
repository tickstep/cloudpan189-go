package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"path"
	"strings"
)

func RunMkdir(familyId int64, name string) {
	activeUser := GetActiveUser()
	fullpath := activeUser.PathJoin(familyId, name)
	pathSlice := strings.Split(fullpath, "/")
	rs := &cloudpan.AppMkdirResult{}
	err := apierror.NewFailedApiError("")

	var cWorkDir = activeUser.Workdir
	var cFileId = activeUser.WorkdirFileEntity.FileId
	if IsFamilyCloud(familyId) {
		cWorkDir = activeUser.FamilyWorkdir
		cFileId = activeUser.FamilyWorkdirFileEntity.FileId
	}
	if path.Dir(fullpath) == cWorkDir {
		rs, err = activeUser.PanClient().AppMkdirRecursive(familyId, cFileId, path.Clean(path.Dir(fullpath)), len(pathSlice) - 1, pathSlice)
	} else {
		rs, err = activeUser.PanClient().AppMkdirRecursive(familyId,"", "", 0, pathSlice)
	}

	if err != nil {
		fmt.Println("创建文件夹失败：" + err.Error())
		return
	}

	if rs.FileId != "" {
		fmt.Println("创建文件夹成功: ", fullpath)
	} else {
		fmt.Println("创建文件夹失败: ", fullpath)
	}
}