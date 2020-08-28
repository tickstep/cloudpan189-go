package panupload

import (
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/library-go/converter"
	"github.com/tickstep/library-go/logger"
)

const (
	// MaxUploadBlockSize 最大上传的文件分片大小
	MaxUploadBlockSize = 2 * converter.GB
	// MinUploadBlockSize 最小的上传的文件分片大小
	MinUploadBlockSize = 4 * converter.MB
	// MaxRapidUploadSize 秒传文件支持的最大文件大小
	MaxRapidUploadSize = 20 * converter.GB

	UploadingFileName = "cloud189_uploading.json"
)

var (
	cmdUploadVerbose = logger.New("CLOUD189_UPLOAD", config.EnvVerbose)
)

func getBlockSize(fileSize int64) int64 {
	blockNum := fileSize / MinUploadBlockSize
	if blockNum > 999 {
		return fileSize/999 + 1
	}
	return MinUploadBlockSize
}
