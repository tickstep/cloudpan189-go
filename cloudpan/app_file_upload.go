package cloudpan

import (
	"encoding/xml"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type (
	// UploadFunc 上传文件处理函数
	UploadFunc func(httpMethod, fullUrl string, headers map[string]string) (resp *http.Response, err error)

	AppCreateUploadFileParam struct {
		// ParentFolderId 存储云盘的目录ID
		ParentFolderId string
		// FileName 存储云盘的文件名
		FileName string
		// Size 文件总大小
		Size int64
		// Md5 文件MD5
		Md5 string
		// LastWrite 文件最后修改日期，格式：2018-11-18 09:12:13
		LastWrite string
		// LocalPath 文件存储的本地绝对路径
		LocalPath string
	}

	AppCreateUploadFileResult struct {
		XMLName xml.Name `xml:"uploadFile"`
		// UploadFileId 上传文件请求ID
		UploadFileId string `xml:"uploadFileId"`
		// FileUploadUrl 上传文件数据的URL路径
		FileUploadUrl string `xml:"fileUploadUrl"`
		// FileCommitUrl 上传文件完成后确认路径
		FileCommitUrl string `xml:"fileCommitUrl"`
		// FileDataExists 文件是否已存在云盘中，0-未存在，1-已存在
		FileDataExists int `xml:"fileDataExists"`
		// 请求的X-Request-ID
		XRequestId string
	}

	AppFileRange struct {
		// 起始值，包含
		Offset int
		// 总上传长度
		Len int
	}

	AppUploadFileCommitResult struct {
		XMLName xml.Name `xml:"file"`
		Id string `xml:"id"`
		Name string `xml:"name"`
		Size string `xml:"size"`
		Md5 string `xml:"md5"`
		CreateDate string `xml:"createDate"`
		Rev string `xml:"rev"`
		UserId string `xml:"userId"`
		RequestId string `xml:"requestId"`
		IsSafe string `xml:"isSafe"`
	}

	AppGetUploadFileStatusResult struct {
		XMLName xml.Name `xml:"uploadFile"`
		// 上传文件的ID
		UploadFileId string `xml:"uploadFileId"`
		// 已上传的大小
		Size int64 `xml:"size"`
		FileUploadUrl string `xml:"fileUploadUrl"`
		FileCommitUrl string `xml:"fileCommitUrl"`
		FileDataExists int `xml:"fileDataExists"`
	}
)

func (p *PanClient) AppCreateUploadFile(param *AppCreateUploadFileParam) (*AppCreateUploadFileResult, *apierror.ApiError) {
	fullUrl := API_URL + "/createUploadFile.action?" + apiutil.PcClientInfoSuffixParam()
	httpMethod := "POST"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := apiutil.XRequestId()
	appToken := p.appToken
	headers := map[string]string {
		"Content-Type": "application/x-www-form-urlencoded",
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl, dateOfGmt),
		"X-Request-ID": requestId,
	}
	formData := map[string]string {
		"parentFolderId": param.ParentFolderId,
		"baseFileId": "",
		"fileName": param.FileName,
		"size": strconv.Itoa(int(param.Size)),
		"md5": param.Md5,
		"lastWrite": param.LastWrite,
		"localPath": strings.ReplaceAll(param.LocalPath, "\\", "/"),
		"opertype": "1",
		"flag": "1",
		"resumePolicy": "1",
		"isLog": "0",
		"fileExt": "",
	}
	logger.Verboseln("do request url: " + fullUrl)
	body, err1 := p.client.Fetch(httpMethod, fullUrl, formData, headers)
	if err1 != nil {
		logger.Verboseln("CreateUploadFile occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	logger.Verboseln("response: " + string(body))
	item := &AppCreateUploadFileResult{}
	if err := xml.Unmarshal(body, item); err != nil {
		logger.Verboseln("CreateUploadFile parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item.XRequestId = requestId
	return item, nil
}

func (p *PanClient) AppUploadFileData(uploadUrl, uploadFileId, xRequestId string, fileRange *AppFileRange, uploadFunc UploadFunc) *apierror.ApiError {
	fullUrl := uploadUrl + "?" + apiutil.PcClientInfoSuffixParam()
	httpMethod := "PUT"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := xRequestId
	appToken := p.appToken
	headers := map[string]string {
		"Content-Type": "application/octet-stream",
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl, dateOfGmt),
		"X-Request-ID": requestId,
		"ResumePolicy": "1",
		"Edrive-UploadFileId": uploadFileId,
		"Edrive-UploadFileRange": "bytes=" + strconv.Itoa(fileRange.Offset) + "-" + strconv.Itoa(fileRange.Len),
		"Expect": "100-continue",
	}
	logger.Verboseln("do request url: " + fullUrl)
	resp, err1 := uploadFunc(httpMethod, fullUrl, headers)
	if err1 != nil {
		logger.Verboseln("AppUploadFileData occurs error: ", err1.Error())
		return apierror.NewApiErrorWithError(err1)
	}
	if resp != nil {
		er := &apierror.AppErrorXmlResp{}
		d, _ := ioutil.ReadAll(resp.Body)
		if err := xml.Unmarshal(d, er); err == nil {
			if er.Code != "" {
				if er.Code == "UploadOffsetVerifyFailed" {
					return apierror.NewApiError(apierror.ApiCodeUploadOffsetVerifyFailed, "上传文件数据偏移值校验失败")
				}
			}
		}
	}
	return nil
}

func (p *PanClient) AppUploadFileCommit(uploadCommitUrl, uploadFileId, xRequestId string) (*AppUploadFileCommitResult, *apierror.ApiError) {
	fullUrl := uploadCommitUrl + "?" + apiutil.PcClientInfoSuffixParam()
	httpMethod := "POST"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := xRequestId
	appToken := p.appToken
	headers := map[string]string {
		"Content-Type": "application/x-www-form-urlencoded",
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl, dateOfGmt),
		"X-Request-ID": requestId,
	}
	formData := map[string]string {
		"uploadFileId": uploadFileId,
		"opertype": "1",
		"ResumePolicy": "1",
		"isLog": "0",
	}
	logger.Verboseln("do request url: " + fullUrl)
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl, formData, headers)
	if err1 != nil {
		logger.Verboseln("AppUploadFileData occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "UploadFileStatusVerifyFailed" {
				return nil, apierror.NewApiError(apierror.ApiCodeUploadFileStatusVerifyFailed, "上传文件校验失败")
			}
		}
	}
	item := &AppUploadFileCommitResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppUploadFileData parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

// AppGetUploadFileStatus 查询上传的文件状态
func (p *PanClient) AppGetUploadFileStatus(uploadFileId string) (*AppGetUploadFileStatusResult, *apierror.ApiError) {
	fullUrl := API_URL + "/getUploadFileStatus.action?uploadFileId=" + uploadFileId + "&ResumePolicy=1&" + apiutil.PcClientInfoSuffixParam()
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	requestId := apiutil.XRequestId()
	appToken := p.appToken
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl, dateOfGmt),
		"X-Request-ID": requestId,
	}
	logger.Verboseln("do request url: " + fullUrl)
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl, nil, headers)
	if err1 != nil {
		logger.Verboseln("AppGetUploadFileStatus occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "UploadFileNotFound" {
				return nil, apierror.NewApiError(apierror.ApiCodeUploadFileNotFound, "服务器上传文件不存在")
			}
		}
	}
	item := &AppGetUploadFileStatusResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppGetUploadFileStatus parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}