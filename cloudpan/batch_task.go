package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"strconv"
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
		TypeFlag BatchTaskType `json:"type"`
		TaskInfos BatchTaskInfoList `json:"taskInfos"`
		TargetFolderId string `json:"targetFolderId"`
		ShareId int64 `json:"shareId"`
	}

    // CheckTaskResult 检查任务结果
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
	BatchTaskType string
)

const (
	// BatchTaskStatusNotAction 无需任何操作
	BatchTaskStatusNotAction BatchTaskStatus = 2
	// BatchTaskStatusOk 成功
	BatchTaskStatusOk BatchTaskStatus = 4

	// BatchTaskTypeDelete 删除文件任务
	BatchTaskTypeDelete BatchTaskType = "DELETE"
	// BatchTaskTypeCopy 复制文件任务
	BatchTaskTypeCopy BatchTaskType = "COPY"
	// BatchTaskTypeMove 移动文件任务
	BatchTaskTypeMove BatchTaskType = "MOVE"

	// BatchTaskTypeRecycleRestore 还原回收站文件
	BatchTaskTypeRecycleRestore BatchTaskType = "RESTORE"

	// BatchTaskTypeShareSave 转录分享
	BatchTaskTypeShareSave BatchTaskType = "SHARE_SAVE"
)

func (p *PanClient) CreateBatchTask (param *BatchTaskParam) (taskId string, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/createBatchTask.action", WEB_URL)
	logger.Verboseln("do request url: " + fullUrl.String())
	taskInfosStr, err := json.Marshal(param.TaskInfos)
	var postData map[string]string
	if BatchTaskTypeDelete == param.TypeFlag || BatchTaskTypeRecycleRestore == param.TypeFlag {
		postData = map[string]string {
			"type": string(param.TypeFlag),
			"taskInfos": string(taskInfosStr),
		}
	} else if BatchTaskTypeCopy == param.TypeFlag || BatchTaskTypeMove == param.TypeFlag {
		postData = map[string]string {
			"type": string(param.TypeFlag),
			"taskInfos": string(taskInfosStr),
			"targetFolderId": param.TargetFolderId,
		}
	} else if BatchTaskTypeShareSave == param.TypeFlag {
		postData = map[string]string {
			"type": string(param.TypeFlag),
			"taskInfos": string(taskInfosStr),
			"targetFolderId": param.TargetFolderId,
			"shareId": strconv.Itoa(int(param.ShareId)),
		}
	} else {
		return "", apierror.NewFailedApiError("不支持的操作")
	}

	body, err := p.client.DoPost(fullUrl.String(), postData)
	if err != nil {
		logger.Verboseln("CreateBatchTask failed")
		return "", apierror.NewApiErrorWithError(err)
	}
	comResp := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, comResp); err == nil {
		if comResp.ErrorCode == "InternalError" {
			logger.Verboseln("response failed", comResp)
			return "", apierror.NewFailedApiError("操作失败")
		}
	}
	return strings.ReplaceAll(string(body), "\"", ""), nil
}

func (p *PanClient) CheckBatchTask (typeFlag BatchTaskType, taskId string) (result *CheckTaskResult, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/checkBatchTask.action", WEB_URL)
	logger.Verboseln("do request url: " + fullUrl.String())
	postData := map[string]string {
		"type": string(typeFlag),
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