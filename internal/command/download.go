package command

import (
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cmder/cmdtable"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/internal/file/downloader"
	"github.com/tickstep/cloudpan189-go/internal/functions/pandownload"
	"github.com/tickstep/cloudpan189-go/internal/taskframework"
	"github.com/tickstep/cloudpan189-go/library/converter"
	"github.com/tickstep/cloudpan189-go/library/requester/transfer"
	"os"
	"path/filepath"
	"runtime"
)

type (
	//DownloadOptions 下载可选参数
	DownloadOptions struct {
		IsPrintStatus        bool
		IsExecutedPermission bool
		IsOverwrite          bool
		SaveTo               string
		Parallel             int
		Load                 int
		MaxRetry             int
		NoCheck              bool
	}

	// LocateDownloadOption 获取下载链接可选参数
	LocateDownloadOption struct {
		FromPan bool
	}
)

var (
	// MaxDownloadRangeSize 文件片段最大值
	MaxDownloadRangeSize = 55 * converter.MB

	// DownloadCacheSize 默认每个线程下载缓存大小
	DownloadCacheSize = 4 * converter.MB
)

func downloadPrintFormat(load int) string {
	if load <= 1 {
		return pandownload.DefaultPrintFormat
	}
	return "\r[%s] ↓ %s/%s %s/s in %s, left %s ..."
}

// RunDownload 执行下载网盘内文件
func RunDownload(paths []string, options *DownloadOptions) {
	if options == nil {
		options = &DownloadOptions{}
	}

	if options.Load <= 0 {
		options.Load = config.Config.MaxDownloadLoad
	}

	if options.MaxRetry < 0 {
		options.MaxRetry = pandownload.DefaultDownloadMaxRetry
	}

	if runtime.GOOS == "windows" {
		// windows下不加执行权限
		options.IsExecutedPermission = false
	}

	// 设置下载配置
	cfg := &downloader.Config{
		Mode:                       transfer.RangeGenMode_BlockSize,
		CacheSize:                  config.Config.CacheSize,
		BlockSize:                  MaxDownloadRangeSize,
		MaxRate:                    config.Config.MaxDownloadRate,
		InstanceStateStorageFormat: downloader.InstanceStateStorageFormatJSON,
	}
	if cfg.CacheSize == 0 {
		cfg.CacheSize = int(DownloadCacheSize)
	}

	// 设置下载最大并发量
	if options.Parallel < 1 {
		options.Parallel = config.Config.MaxDownloadParallel
	}

	paths, err := matchPathByShellPattern(paths...)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("\n")
	fmt.Printf("[0] 提示: 当前下载最大并发量为: %d, 下载缓存为: %d\n", options.Parallel, cfg.CacheSize)

	var (
		panClient = GetActivePanClient()
		loadCount = 0
	)

	// 预测要下载的文件数量
	for k := range paths {
		// 使用递归获取文件的方法计算路径包含的文件的总数量
		panClient.FilesDirectoriesRecurseList(paths[k], func(depth int, _ string, fd *cloudpan.FileEntity, apiError *apierror.ApiError) bool {
			if apiError != nil {
				panCommandVerbose.Warnf("%s\n", apiError)
				return true
			}

			// 忽略统计文件夹数量
			if !fd.IsFolder {
				loadCount++
				if loadCount >= options.Load { // 文件的总数量超过指定的指定数量，则不再进行下层的递归查找文件
					return false
				}
			}
			return true
		})

		if loadCount >= options.Load {
			break
		}
	}

	// 修改Load, 设置MaxParallel
	if loadCount > 0 {
		options.Load = loadCount
		// 取平均值
		cfg.MaxParallel = config.AverageParallel(options.Parallel, loadCount)
	} else {
		cfg.MaxParallel = options.Parallel
	}

	var (
		executor = taskframework.TaskExecutor{
			IsFailedDeque: true, // 统计失败的列表
		}
		statistic = &pandownload.DownloadStatistic{}
	)
	// 处理队列
	for k := range paths {
		newCfg := *cfg
		unit := pandownload.DownloadTaskUnit{
			Cfg:                  &newCfg, // 复制一份新的cfg
			PanClient:            panClient,
			VerbosePrinter:       panCommandVerbose,
			PrintFormat:          downloadPrintFormat(options.Load),
			ParentTaskExecutor:   &executor,
			DownloadStatistic:    statistic,
			IsPrintStatus:        options.IsPrintStatus,
			IsExecutedPermission: options.IsExecutedPermission,
			IsOverwrite:          options.IsOverwrite,
			NoCheck:              options.NoCheck,
			FilePanPath:          paths[k],
		}

		// 设置储存的路径
		if options.SaveTo != "" {
			unit.SavePath = filepath.Join(options.SaveTo, filepath.Base(paths[k]))
		} else {
			// 使用默认的保存路径
			unit.SavePath = GetActiveUser().GetSavePath(paths[k])
		}
		info := executor.Append(&unit, options.MaxRetry)
		fmt.Printf("[%s] 加入下载队列: %s\n", info.Id(), paths[k])
	}

	// 开始计时
	statistic.StartTimer()

	// 开始执行
	executor.Execute()

	fmt.Printf("\n下载结束, 时间: %s, 数据总量: %s\n", statistic.Elapsed()/1e6*1e6, converter.ConvertFileSize(statistic.TotalSize()))

	// 输出失败的文件列表
	failedList := executor.FailedDeque()
	if failedList.Size() != 0 {
		fmt.Printf("以下文件下载失败: \n")
		tb := cmdtable.NewTable(os.Stdout)
		for e := failedList.Shift(); e != nil; e = failedList.Shift() {
			item := e.(*taskframework.TaskInfoItem)
			tb.Append([]string{item.Info.Id(), item.Unit.(*pandownload.DownloadTaskUnit).FilePanPath})
		}
		tb.Render()
	}
}
