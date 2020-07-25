package pandownload

import (
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"os"
)

// CheckFileValid 检测文件有效性
func CheckFileValid(filePath string, fileInfo *cloudpan.FileEntity) error {
	// 检查MD5
	// 检查文件大小
	// 检查digest签名
	return nil
}

// FileExist 检查文件是否存在,
// 只有当文件存在, 文件大小不为0或断点续传文件不存在时, 才判断为存在
func FileExist(path string) bool {
	if info, err := os.Stat(path); err == nil {
		if info.Size() == 0 {
			return false
		}
		if _, err = os.Stat(path + DownloadSuffix); err != nil {
			return true
		}
	}

	return false
}
