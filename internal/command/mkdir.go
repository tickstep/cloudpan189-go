package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"path"
	"strings"
)

func RunMkdir(name string) {
	activeUser := GetActiveUser()
	fullpath := activeUser.PathJoin(name)
	pathSlice := strings.Split(fullpath, "/")
	rs := &cloudpan.MkdirResult{}
	err := apierror.NewFailedApiError("")
	if path.Dir(fullpath) == activeUser.Workdir {
		rs, err = activeUser.PanClient().MkdirRecursive(activeUser.WorkdirFileEntity.FileId, path.Clean(path.Dir(fullpath)), len(pathSlice) - 1, pathSlice)
	} else {
		rs, err = activeUser.PanClient().MkdirRecursive("", "", 0, pathSlice)
	}

	if err != nil {
		fmt.Println("创建文件夹失败：" + err.Error())
		return
	}

	if rs.IsNew {
		fmt.Println("创建文件夹成功: ", fullpath)
	} else {
		if rs.FileId != "" {
			fmt.Println("文件夹已存在: ", fullpath)
		} else {
			fmt.Println("创建文件夹失败: ", fullpath)
		}
	}
}