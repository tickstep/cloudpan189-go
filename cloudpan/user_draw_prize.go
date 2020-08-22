package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"strings"
)

type (
	userDrawPrizeResp struct {
		ActivityId string `json:"activityId"`
		Description string `json:"description"`
		IsUsed int `json:"isUsed"`
		ListId int `json:"listId"`
		PrizeGrade int `json:"prizeGrade"`
		PrizeId string `json:"prizeId"`
		PrizeName string `json:"prizeName"`
		PrizeStatus int `json:"prizeStatus"`
		PrizeType int `json:"prizeType"`
		UseDate string `json:"useDate"`
		UserId int64 `json:"userId"`
	}

	UserDrawPrizeResult struct {
		Success bool
		Tip string
	}

	ActivityTaskId string
)

const (
	ActivitySignin ActivityTaskId = "TASK_SIGNIN"
	ActivitySignPhotos ActivityTaskId = "TASK_SIGNIN_PHOTOS"
)

// 抽奖
func (p *PanClient) UserDrawPrize(taskId ActivityTaskId) (*UserDrawPrizeResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "https://m.cloud.189.cn/v2/drawPrizeMarketDetails.action?taskId=%s&activityId=ACT_SIGNIN",
		taskId)
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		return nil, apierror.NewApiErrorWithError(err)
	}

	item := &userDrawPrizeResp{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("UserDrawPrize parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}

	result := UserDrawPrizeResult{}
	if item.PrizeStatus == 1 {
		result.Success = true
		result.Tip = item.Description
		return &result, nil
	}
	return nil, apierror.NewFailedApiError("抽奖失败")
}