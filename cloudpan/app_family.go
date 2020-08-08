package cloudpan

import (
	"encoding/xml"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"strings"
)

type (
	// AppGetFileInfoParam 获取文件信息参数
	AppFamilyInfo struct {
		Count int `xml:"count"`
		Type int `xml:"type"`
		UserRole int `xml:"userRole"`
		CreateTime string `xml:"createTime"`
		FamilyId int64 `xml:"familyId"`
		RemarkName string `xml:"remarkName"`
		UseFlag int `xml:"useFlag"`
	}

	AppFamilyInfoListResult struct {
		XMLName xml.Name `xml:"familyListResponse"`
		FamilyInfoList []AppFamilyInfo `xml:"familyInfo"`
	}

)

// AppGetFamilyList 获取用户的家庭列表
func (p *PanClient) AppGetFamilyList() (*AppFamilyInfoListResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/family/manage/getFamilyList.action?%s",
		API_URL, apiutil.PcClientInfoSuffixParam())
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	appToken := p.appToken
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.FamilySessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.FamilySessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppGetFamilyList occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "FamilyOperationFailed" {
			return nil, apierror.NewFailedApiError("获取家庭列表错误")
		}
	}
	item := &AppFamilyInfoListResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppGetFamilyList parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}