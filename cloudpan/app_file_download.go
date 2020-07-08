package cloudpan

import (
	"encoding/xml"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"net/http"
	"strconv"
	"strings"
)

type (

)

func (p *PanClient) AppGetFileDownloadUrl(fileId string) (string, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	appToken := p.appToken
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	fmt.Fprintf(fullUrl, "%s/getFileDownloadUrl.action?fileId=%s&dt=3&flag=1&%s",
		API_URL, fileId, apiutil.PcClientInfoSuffixParam())
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppGetFileDownloadUrl occurs error: ", err1.Error())
		return "", apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))

	type fdUrl struct {
		XMLName xml.Name `xml:"fileDownloadUrl"`
		FileDownloadUrl string `xml:",innerxml"`
	}

	item := &fdUrl{}
	if err := xml.Unmarshal(body, item); err != nil {
		fmt.Println("AppGetFileDownloadUrl parse response failed")
		return "", apierror.NewApiErrorWithError(err)
	}
	return strings.ReplaceAll(item.FileDownloadUrl, "&amp;", "&"), nil
}

func (p *PanClient) AppDownloadFileData(downloadFileUrl string, fileRange AppFileRange) (resp *http.Response, error *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	appToken := p.appToken
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	fmt.Fprintf(fullUrl, "%s&%s",
		downloadFileUrl, apiutil.PcClientInfoSuffixParam())
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}
	// 支持断点续传
	if fileRange.Start != 0 || fileRange.End != 0 {
		rangeStr := "bytes=" + strconv.Itoa(fileRange.Start) + "-"
		if fileRange.End != 0 {
			rangeStr += strconv.Itoa(fileRange.End)
		}
		headers["range"] = rangeStr
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	resp, err := p.client.Req(httpMethod, fullUrl.String(), nil, headers)
	if err != nil {
		logger.Verboseln("AppDownloadFileData response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return resp, nil
}