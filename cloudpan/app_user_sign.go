package cloudpan

import (
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
)

type (
	AppUserSignResult struct {
		Success bool
		tip string
	}
)

// AppUserSign 用户签到
func (p *PanClient) AppUserSign(appToken *AppLoginToken) (*AppUserSignResult, *apierror.ApiError) {
	//p.client.SetProxy("http://127.0.0.1:8888")
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
	xmlStr := string(body)
	result.tip = xmlStr
	return &result, nil
}
