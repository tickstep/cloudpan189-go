package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"path"
	"strconv"
	"strings"
)

type (
	// HandleFileDirectoryFunc 处理文件或目录的元信息, 返回值控制是否退出递归
	HandleFileDirectoryFunc func(depth int, fdPath string, fd *FileEntity, apierr *apierror.ApiError) bool

	MediaType uint
	OrderBy uint
	OrderSort string

	// FileListParam 文件列表参数
	FileListParam struct {
		// FileId 文件ID
		FileId string
		// MediaType 媒体文件过滤
		MediaType MediaType
		// InGroupSpace ???
		InGroupSpace bool
		// OrderBy 排序字段
		OrderBy OrderBy
		// OrderSort 排序顺序
		OrderSort OrderSort
		// PageNum 页数量，从1开始
		PageNum uint
		// PageSize 页大小，默认60
		PageSize uint
	}

	// FileSearchParam 文件搜索参数
	FileSearchParam struct {
		FileListParam
		// FileId 文件ID
		// Keyword 搜索关键字
		Keyword string
	}

	FileList []*FileEntity
	PathList []*PathEntity

	// FileSearchResult 文件搜索返回结果
	FileSearchResult struct {
		// Data 数据
		Data FileList `json:"data"`
		// PageNum 页数量，从1开始
		PageNum uint `json:"pageNum"`
		// PageSize 页大小，默认60
		PageSize uint `json:"pageSize"`
		// Path 路径
		Path PathList `json:"path"`
		// RecordCount 文件总数量
		RecordCount uint `json:"recordCount"`
	}

	FileEntity struct {
		// CreateTime 创建时间
		CreateTime string `json:"createTime"`
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileIdDigest 文件ID指纹
		FileIdDigest string `json:"fileIdDigest"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// FileSize 文件大小，文件夹为0
		FileSize int64 `json:"fileSize"`
		// FileType 文件类型，后缀名，例如:"dmg"，没有则为空
		FileType string `json:"fileType"`
		// IsFolder 是否是文件夹
		IsFolder bool `json:"isFolder"`
		// LastOpTime 最后修改时间
		LastOpTime string `json:"lastOpTime"`
		// ParentId 父文件ID
		ParentId string `json:"parentId"`

		// DownloadUrl 下载路径，只有文件才有
		DownloadUrl string `json:"downloadUrl"`
		// IsStarred 是否是星标文件
		IsStarred bool `json:"isStarred"`
		// MediaType 媒体类型
		MediaType MediaType `json:"mediaType"`
		// SubFileCount 文件夹子文件数量，对文件夹详情有效
		SubFileCount uint `json:"subFileCount"`
		// FilePath 文件的完整路径
		Path string `json:"path"`
	}

	PathEntity struct {
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// IsCoShare ???
		IsCoShare uint `json:"isCoShare"`
	}


)

// NewFileSearchParam 创建默认搜索参数
func NewFileSearchParam() *FileSearchParam {
	flp := NewFileListParam()
	return &FileSearchParam {
		FileListParam{
			FileId: flp.FileId,
			MediaType: flp.MediaType,
			InGroupSpace: flp.InGroupSpace,
			OrderBy: flp.OrderBy,
			OrderSort: flp.OrderSort,
			PageNum: flp.PageNum,
			PageSize: flp.PageSize,
		},
		"",
	}
}

func NewFileListParam() *FileListParam {
	return &FileListParam {
		FileId: "-11",
		MediaType: MediaTypeDefault,
		InGroupSpace: false,
		OrderBy: OrderByName,
		OrderSort: OrderAsc,
		PageNum: 1,
		PageSize: 60,
	}
}

// NewFileEntityForRootDir 创建根目录"/"的默认文件信息
func NewFileEntityForRootDir() *FileEntity {
	return &FileEntity {
		FileId: "-11",
		IsFolder: true,
		FileName: "/",
		ParentId: "",
	}
}

const (
	// MediaTypeDefault 默认全部
	MediaTypeDefault MediaType = 0
	// MediaTypeMusic 音乐
	MediaTypeMusic MediaType = 1
	// MediaTypeVideo 视频
	MediaTypeVideo MediaType = 3
	// MediaTypeDocument 文档
	MediaTypeDocument MediaType = 4

	// OrderByName 文件名
	OrderByName OrderBy = 1
	// OrderBySize 大小
	OrderBySize OrderBy = 2
	// OrderByTime 时间
	OrderByTime OrderBy = 3

	// OrderAsc 升序
	OrderAsc OrderSort = "ASC"
	// OrderDesc 降序
	OrderDesc OrderSort = "DESC"
)



// TotalSize 获取目录下文件的总大小
func (fl FileList) TotalSize() int64 {
	var size int64
	for k := range fl {
		if fl[k] == nil {
			continue
		}

		size += fl[k].FileSize
	}
	return size
}


// Count 获取文件总数和目录总数
func (fl FileList) Count() (fileN, directoryN int64) {
	for k := range fl {
		if fl[k] == nil {
			continue
		}

		if fl[k].IsFolder {
			directoryN++
		} else {
			fileN++
		}
	}
	return
}
func (f *FileEntity) String() string {
	builder := &strings.Builder{}
	builder.WriteString("文件ID: " + f.FileId + "\n")
	builder.WriteString("文件名: " + f.FileName + "\n")
	if f.IsFolder {
		builder.WriteString("文件类型: 目录\n")
	} else {
		builder.WriteString("文件类型: 文件\n")
	}
	builder.WriteString("文件路径: " + f.Path + "\n")
	return builder.String()
}

func (p *PanClient) FileList(param *FileListParam) (result *FileSearchResult, error *apierror.ApiError) {
	fsp := NewFileSearchParam()
	fsp.FileId = param.FileId
	fsp.MediaType = param.MediaType
	fsp.InGroupSpace = param.InGroupSpace
	fsp.OrderBy = param.OrderBy
	fsp.OrderSort = param.OrderSort
	fsp.PageNum = param.PageNum
	fsp.PageSize = param.PageSize

	// 搜索中有keyword，那么FileSearchResult.path路径就不对了
	fsp.Keyword = ""

	item, er := p.FileSearch(fsp)
	if er != nil {
		return nil, er
	}

	// combine the path for file
	parentDirPath := strings.Builder{}
	for _, p := range item.Path {
		if p.FileName == "全部文件" {
			parentDirPath.WriteString("/")
			continue
		}
		parentDirPath.WriteString(p.FileName + "/")
	}
	pd := parentDirPath.String()

	// add path to file
	for _, f := range item.Data {
		f.Path = pd + f.FileName
	}

	return item, nil
}

func (p *PanClient) FileSearch(param *FileSearchParam) (result *FileSearchResult, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	md := ""
	if param.MediaType != 0 {
		md = strconv.Itoa(int(param.MediaType))
	}
	fmt.Fprintf(fullUrl, "%s/v2/listFiles.action?fileId=%s&mediaType=%s&keyword=%s&inGroupSpace=%t&orderBy=%d&order=%s&pageNum=%d&pageSize=%d",
		WEB_URL, param.FileId, md, param.Keyword, param.InGroupSpace, param.OrderBy, param.OrderSort,
		param.PageNum, param.PageSize)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("search failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &FileSearchResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("search response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

func (p *PanClient) FileInfoById(fileId string) (fileInfo *FileEntity, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/getFileInfo.action?fileId=%s", WEB_URL, fileId)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("get file info failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &FileEntity{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("file info response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

// FileInfoByPath 通过路径获取文件详情，pathStr是绝对路径
func (p *PanClient) FileInfoByPath(pathStr string) (fileInfo *FileEntity, error *apierror.ApiError) {
	if pathStr == "" {
		pathStr = "/"
	}
	//pathStr = path.Clean(pathStr)
	if !path.IsAbs(pathStr) {
		return nil, apierror.NewFailedApiError("pathStr必须是绝对路径")
	}
	if len(pathStr) > 1 {
		pathStr = path.Clean(pathStr)
	}

	var pathSlice []string
	if pathStr == "/" {
		pathSlice = []string{""}
	} else {
		pathSlice = strings.Split(pathStr, PathSeparator)
		if pathSlice[0] != "" {
			return nil, apierror.NewFailedApiError("pathStr必须是绝对路径")
		}
	}
	return p.getFileInfoByPath(0, &pathSlice, nil)
}

func (p *PanClient) getFileInfoByPath(index int, pathSlice *[]string, parentFileInfo *FileEntity) (*FileEntity, *apierror.ApiError)  {
	if parentFileInfo == nil {
		// default root "/" entity
		parentFileInfo = NewFileEntityForRootDir()
		if index == 0 && len(*pathSlice) == 1 {
			// root path "/"
			return parentFileInfo, nil
		}

		return p.getFileInfoByPath(index + 1, pathSlice, parentFileInfo)
	}

	if index >= len(*pathSlice) {
		return parentFileInfo, nil
	}

	searchPath := NewFileListParam()
	searchPath.FileId = parentFileInfo.FileId
	fileResult, err := p.FileList(searchPath)
	if err != nil {
		return nil, err
	}

	if fileResult == nil || fileResult.Data == nil || len(fileResult.Data) == 0 {
		return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
	}
	for _, fileEntity := range fileResult.Data {
		if fileEntity.FileName == (*pathSlice)[index] {
			return p.getFileInfoByPath(index + 1, pathSlice, fileEntity)
		}
	}
	return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
}

// FilesDirectoriesRecurseList 递归获取目录下的文件和目录列表
func (p *PanClient) FilesDirectoriesRecurseList(path string, handleFileDirectoryFunc HandleFileDirectoryFunc) FileList {
	targetFileInfo, er := p.FileInfoByPath(path)
	if er != nil {
		if handleFileDirectoryFunc != nil {
			handleFileDirectoryFunc(0, path, nil, er)
		}
		return nil
	}
	if !targetFileInfo.IsFolder {
		if handleFileDirectoryFunc != nil {
			handleFileDirectoryFunc(0, path, targetFileInfo, nil)
		}
		return FileList{targetFileInfo}
	}

	fld := &FileList{}
	ok := p.recurseList(targetFileInfo, 1, handleFileDirectoryFunc, fld)
	if !ok {
		return nil
	}
	return *fld
}

func (p *PanClient) recurseList(folderInfo *FileEntity, depth int, handleFileDirectoryFunc HandleFileDirectoryFunc, fld *FileList) bool {
	flp := NewFileListParam()
	flp.FileId = folderInfo.FileId
	r, apiError := p.FileList(flp)
	if apiError != nil {
		if handleFileDirectoryFunc != nil {
			handleFileDirectoryFunc(depth, folderInfo.Path, nil, apiError) // 传递错误
		}
		return false
	}
	ok := true
	for _, fi := range r.Data {
		*fld = append(*fld, fi)
		if fi.IsFolder {
			ok = p.recurseList(fi, depth+1, handleFileDirectoryFunc, fld)
		} else {
			if handleFileDirectoryFunc != nil {
				ok = handleFileDirectoryFunc(depth, fi.Path, fi, nil)
			}
		}
		if !ok {
			return false
		}
	}
	return true
}