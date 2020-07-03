package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/internal/config"
)

func RunChangeDirectory(targetPath string) {
	user := config.Config.ActiveUser()
	targetPath = user.PathJoin(targetPath)

	targetPathInfo, err := user.PanClient().FileInfoByPath(targetPath)
	if err != nil {
		fmt.Println(err)
		return
	}

	if !targetPathInfo.IsFolder {
		fmt.Printf("错误: %s 不是一个目录 (文件夹)\n", targetPath)
		return
	}

	user.Workdir = targetPath
	user.WorkdirFileEntity = *targetPathInfo

	fmt.Printf("改变工作目录: %s\n", targetPath)
}
