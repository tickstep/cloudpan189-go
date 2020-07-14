package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"path"
)



// GetFileInfoByPaths 获取指定文件路径的文件详情信息
func GetFileInfoByPaths(paths ...string) (fileInfoList []*cloudpan.FileEntity, failedPaths []string, error error) {
	if len(paths) <= 0 {
		return nil, nil, fmt.Errorf("请指定文件路径")
	}
	activeUser := GetActiveUser()

	for idx := 0; idx < len(paths); idx++ {
		absolutePath := path.Clean(activeUser.PathJoin(paths[idx]))
		fe, err := activeUser.PanClient().FileInfoByPath(absolutePath)
		if err != nil {
			failedPaths = append(failedPaths, absolutePath)
			continue
		}
		fileInfoList = append(fileInfoList, fe)
	}
	return
}