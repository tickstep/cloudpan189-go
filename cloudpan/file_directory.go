package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"strings"
)

type (
	MediaType uint
	OrderBy uint
	OrderSort string

	// FileSearchParam 文件搜索参数
	FileSearchParam struct {
		// 文件ID
		FileId string
		// 媒体文件过滤
		MediaType MediaType
		// 搜索关键字
		Keyword string
		// ???
		InGroupSpace bool
		// 排序字段
		OrderBy OrderBy
		// 排序顺序
		OrderSort OrderSort
		// 页数量，从1开始
		PageNum uint
		// 页大小，默认60
		PageSize uint
	}

	// FileSearchResult 文件搜索返回结果
	FileSearchResult struct {
		// Data 数据
		Data []*FileEntity `json:"data"`
		// 页数量，从1开始
		PageNum uint `json:"pageNum"`
		// 页大小，默认60
		PageSize uint `json:"pageSize"`
		// Path 路径
		Path []*PathEntity `json:"path"`
		// RecordCount 文件总数量
		RecordCount uint `json:"recordCount"`
	}

	FileEntity struct {
		// CreateTime 创建时间
		CreateTime string `json:"createTime"`
		// DownloadUrl 下载路径，只有文件才有
		DownloadUrl string `json:"downloadUrl"`
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileIdDigest 文件ID指纹
		FileIdDigest string `json:"fileIdDigest"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// FileSize 文件大小，文件夹为0
		FileSize uint64 `json:"fileSize"`
		// FileType 文件类型，后缀名，例如:"dmg"
		FileType string `json:"fileType"`
		// IsFolder 是否是文件夹
		IsFolder bool `json:"isFolder"`
		// IsStarred 是否是星标文件
		IsStarred bool `json:"isStarred"`
		// LastOpTime 最后修改时间
		LastOpTime string `json:"lastOpTime"`
		// MediaType 媒体类型
		MediaType MediaType `json:"mediaType"`
		// ParentId 父文件ID
		ParentId string `json:"parentId"`
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
	return &FileSearchParam{
		FileId: "-11",
		MediaType: MediaTypeDefault,
		InGroupSpace: false,
		OrderBy: OrderByName,
		OrderSort: OrderAsc,
		PageNum: 1,
		PageSize: 60,
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

func (p *PanClient) FileSearch(param *FileSearchParam) (result *FileSearchResult, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/listFiles.action?fileId=%s&mediaType=%s&keyword=%s&inGroupSpace=%b&orderBy=%d&order=%s&pageNum=%d&pageSize=%d",
		WEB_URL, param.FileId, param.MediaType, param.Keyword, param.InGroupSpace, param.OrderBy, param.OrderSort,
		param.PageNum, param.PageSize)
	logger.Verboseln("do reqeust url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("search failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &FileSearchResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("heartbeat response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}
