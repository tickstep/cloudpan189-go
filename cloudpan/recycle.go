package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"net/url"
	"strings"
)

type (
	// RecycleFileInfo 回收站中文件/目录信息
	RecycleFileInfo struct {
		// CreateTime 创建时间
		CreateTime string `json:"createTime"`
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// FileSize 文件大小，文件夹为0
		FileSize int64 `json:"fileSize"`
		// FileType 文件类型，后缀名，例如:"dmg"，没有则为空
		FileType string `json:"fileType"`
		// IsFolder 是否是文件夹
		IsFolder bool `json:"isFolder"`
		// IsFamilyFile 是否是家庭云文件
		IsFamilyFile bool `json:"isFamilyFile"`
		// LastOpTime 最后修改时间
		LastOpTime string `json:"lastOpTime"`
		// ParentId 父文件ID
		ParentId string `json:"parentId"`
		// DownloadUrl 下载路径，只有文件才有
		DownloadUrl string `json:"downloadUrl"`
		// MediaType 媒体类型
		MediaType MediaType `json:"mediaType"`
		// PathStr 文件的完整路径
		PathStr string `json:"pathStr"`
	}

	RecycleFileInfoList []*RecycleFileInfo

	RecycleFileListResult struct {
		// Data 数据
		Data RecycleFileInfoList `json:"data"`
		// PageNum 页数量，从1开始
		PageNum uint `json:"pageNum"`
		// PageSize 页大小，默认60
		PageSize uint `json:"pageSize"`
		// RecordCount 文件总数量
		RecordCount uint `json:"recordCount"`
		FamilyId int `json:"familyId"`
		FamilyName string `json:"familyName"`
	}

	RecycleFileActResult struct {
		Success bool `json:"success"`
	}
)

// RecycleList 列出回收站文件列表
func (p *PanClient) RecycleList(pageNum, pageSize int) (result *RecycleFileListResult, error *apierror.ApiError) {
	if pageNum <= 1 {
		pageNum = 1
	}
	if pageSize <= 1 {
		pageSize = 60
	}
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/listRecycleBin.action?pageNum=%d&pageSize=%d",
		WEB_URL, pageNum, pageSize)
	logger.Verboseln("do request url: " + fullUrl.String())
	//header := map[string]string {
	//	"X-Requested-With": "XMLHttpRequest",
	//	"Accept": "*/*",
	//	"Referer": "https://cloud.189.cn/main.action",
	//}
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("RecycleList failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &RecycleFileListResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("RecycleList response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

// RecycleDelete 删除回收站文件或目录
func (p *PanClient) RecycleDelete(familyId int, fileIdList []string) *apierror.ApiError {
	fullUrl := &strings.Builder{}
	if fileIdList == nil {
		return nil
	}
	if familyId <=0 {
		fmt.Fprintf(fullUrl, "%s/v2/deleteFile.action?fileIdList=%s",
			WEB_URL, url.QueryEscape(strings.Join(fileIdList, ",")))
	} else {
		fmt.Fprintf(fullUrl, "%s/v2/deleteFile.action?familyId=%d&fileIdList=%s",
			WEB_URL, familyId, url.QueryEscape(strings.Join(fileIdList, ",")))
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("RecycleDelete failed")
		return apierror.NewApiErrorWithError(err)
	}
	item := &RecycleFileActResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("RecycleDelete response failed")
		return apierror.NewApiErrorWithError(err)
	}
	if !item.Success {
		return apierror.NewFailedApiError("failed")
	}
	return nil
}

func (p *PanClient) RecycleRestore(fileList []*RecycleFileInfo) (taskId string, err *apierror.ApiError) {
	if fileList == nil {
		return "", nil
	}

	taskReqParam := &BatchTaskParam{
		TypeFlag: BatchTaskTypeRecycleRestore,
		TaskInfos: makeBatchTaskInfoList(fileList),
	}
	return p.CreateBatchTask(taskReqParam)
}

func makeBatchTaskInfoList(opFileList []*RecycleFileInfo) (infoList BatchTaskInfoList) {
	for _, fe := range opFileList {
		isFolder := 0
		if fe.IsFolder {
			isFolder = 1
		}
		infoItem := &BatchTaskInfo{
			FileId: fe.FileId,
			FileName: fe.FileName,
			IsFolder: isFolder,
			SrcParentId: fe.ParentId,
		}
		infoList = append(infoList, infoItem)
	}
	return
}

func (p *PanClient) RecycleClear(familyId int) *apierror.ApiError {
	fullUrl := &strings.Builder{}
	if familyId <=0 {
		fmt.Fprintf(fullUrl, "%s/v2/emptyRecycleBin.action",
			WEB_URL)
	} else {
		fmt.Fprintf(fullUrl, "%s/v2/emptyRecycleBin.action?familyId=%d",
			WEB_URL, familyId)
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("RecycleClear failed")
		return apierror.NewApiErrorWithError(err)
	}
	item := &RecycleFileActResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("RecycleClear response failed")
		return apierror.NewApiErrorWithError(err)
	}
	if !item.Success {
		return apierror.NewFailedApiError("failed")
	}
	return nil
}