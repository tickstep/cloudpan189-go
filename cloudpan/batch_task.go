package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"strings"
)

type (
	// TaskInfo 任务信息
	BatchTaskInfo struct {
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// IsFolder 是否是文件夹，0-否，1-是
		IsFolder int `json:"isFolder"`
		// SrcParentId 文件所在父目录ID
		SrcParentId string `json:"srcParentId"`
	}

	BatchTaskInfoList []*BatchTaskInfo

	// BatchTaskParam 任务参数
	BatchTaskParam struct {
		TypeFlag string `json:"type"`
		TaskInfos BatchTaskInfoList `json:"taskInfos"`
	}

	CheckTaskResult struct {
		FailedCount int `json:"failedCount"`
		SkipCount int `json:"skipCount"`
		SubTaskCount int `json:"subTaskCount"`
		SuccessedCount int `json:"successedCount"`
		SuccessedFileIdList []int64 `json:"successedFileIdList"`
		TaskId string `json:"taskId"`
		// TaskStatus 任务状态， 4-成功
		TaskStatus BatchTaskStatus `json:"taskStatus"`
	}

	BatchTaskStatus int
)

const (
	// BatchTaskStatusOk 成功
	BatchTaskStatusOk = 4
)

func (p *PanClient) CreateBatchTask (param *BatchTaskParam) (taskId string, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/createBatchTask.action", WEB_URL)
	logger.Verboseln("do reqeust url: " + fullUrl.String())
	taskInfosStr, err := json.Marshal(param.TaskInfos)
	postData := map[string]string {
		"type": param.TypeFlag,
		"taskInfos": string(taskInfosStr),
	}
	body, err := p.client.DoPost(fullUrl.String(), postData)
	if err != nil {
		logger.Verboseln("CreateBatchTask failed")
		return "", apierror.NewApiErrorWithError(err)
	}
	return strings.ReplaceAll(string(body), "\"", ""), nil
}

func (p *PanClient) CheckBatchTask (typeFlag, taskId string) (result *CheckTaskResult, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/checkBatchTask.action", WEB_URL)
	logger.Verboseln("do reqeust url: " + fullUrl.String())
	postData := map[string]string {
		"type": typeFlag,
		"taskId": taskId,
	}
	body, err := p.client.DoPost(fullUrl.String(), postData)
	if err != nil {
		logger.Verboseln("CheckBatchTask failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &CheckTaskResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("CheckBatchTask response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}