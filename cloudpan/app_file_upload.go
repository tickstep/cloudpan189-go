package cloudpan

import (
	"encoding/xml"
	"github.com/satori/go.uuid"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"strconv"
	"strings"
)

type (
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
		FileDataExists string `xml:"fileDataExists"`
		// 请求的X-Request-ID
		XRequestId string
	}
)

func (p *PanClient) AppCreateUploadFile(param *AppCreateUploadFileParam) (result *AppCreateUploadFileResult, error *apierror.ApiError) {
	fullUrl := API_URL + "/createUploadFile.action?" + apiutil.PcClientInfoSuffixParam()
	httpMethod := "POST"
	dateOfGmt := apiutil.DateOfGmtStr()
	u4 := uuid.NewV4()
	requestId := strings.ToUpper(u4.String())
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