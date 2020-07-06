package cloudpan

import (
	"encoding/xml"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
)

type (
	AppUserSignStatus int
	AppUserSignResult struct {
		Status AppUserSignStatus
		Tip string
	}

	userSignResult struct {
		XMLName xml.Name `xml:"userSignResult"`
		Result int `xml:"result"`
		ResultTip string `xml:"resultTip"`
		ActivityFlag int `xml:"activityFlag"`
		PrizeListUrl string `xml:"prizeListUrl"`
		ButtonTip string `xml:"buttonTip"`
		ButtonUrl string `xml:"buttonUrl"`
		ActivityTip string `xml:"activityTip"`
	}
)

const (
	AppUserSignStatusFailed AppUserSignStatus = 0
	AppUserSignStatusSuccess AppUserSignStatus = 1
	AppUserSignStatusHasSign AppUserSignStatus = -1
)

// AppUserSign 用户签到
func (p *PanClient) AppUserSign(appToken *AppLoginToken) (*AppUserSignResult, *apierror.ApiError) {
	result := AppUserSignResult{}
	fullUrl := API_URL + "//mkt/userSign.action"
	headers := map[string]string {
		"SessionKey": appToken.SessionKey,
	}
	logger.Verboseln("do request url: " + fullUrl)
	body, err1 := appClient.Fetch("GET", fullUrl, nil, headers)
	if err1 != nil {
		logger.Verboseln("AppUserSign occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))
	item := &userSignResult{}
	if err := xml.Unmarshal(body, item); err != nil {
		logger.Verboseln("AppUserSign parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	switch item.Result {
	case 1:
		result.Status = AppUserSignStatusSuccess
		break
	case -1:
		result.Status = AppUserSignStatusHasSign
		break
	default:
		result.Status = AppUserSignStatusFailed
	}
	result.Tip = item.ResultTip
	return &result, nil
}
