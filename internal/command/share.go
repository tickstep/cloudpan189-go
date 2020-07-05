package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
)

// RunShareSet 执行分享
func RunShareSet(paths []string, expiredTime cloudpan.ShareExpiredTime) {
	fileList, _, err := GetFileInfoByPaths(paths[:len(paths)]...)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, fi := range fileList {
		r, err := GetActivePanClient().PrivateShare(fi.FileId, expiredTime)
		if err != nil {
			fmt.Printf("创建分享链接失败: %s - %s\n", fi.Path, err)
			continue
		}
		fmt.Printf("路径: %s 链接: %s（访问码：%s）\n",fi.Path, r.ShortShareUrl, r.AccessCode)
	}
}

