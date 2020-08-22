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
	DownloadFuncCallback func(httpMethod, fullUrl string, headers map[string]string) (resp *http.Response, err error)

	AppFileDownloadRange struct {
		// 起始值，包含
		Offset int64
		// 结束值，包含
		End int64
	}
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

func (p *PanClient) AppDownloadFileData(downloadFileUrl string, fileRange AppFileDownloadRange, downloadFunc DownloadFuncCallback) *apierror.ApiError {
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
	if fileRange.Offset != 0 || fileRange.End != 0 {
		rangeStr := "bytes=" + strconv.FormatInt(fileRange.Offset, 10) + "-"
		if fileRange.End != 0 {
			rangeStr += strconv.FormatInt(fileRange.End, 10)
		}
		headers["range"] = rangeStr
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	_, err := downloadFunc(httpMethod, fullUrl.String(), headers)
	//resp, err := p.client.Req(httpMethod, fullUrl.String(), nil, headers)
	if err != nil {
		logger.Verboseln("AppDownloadFileData response failed")
		return apierror.NewApiErrorWithError(err)
	}
	return nil
}