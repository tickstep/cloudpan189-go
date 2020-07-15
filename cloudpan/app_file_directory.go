package cloudpan

import (
	"encoding/xml"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/cloudpan/apiutil"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"net/url"
	"strings"
)

type (
	// AppGetFileInfoParam 获取文件信息参数
	AppGetFileInfoParam struct {
		// FileId 文件ID，支持文件和文件夹
		FileId   string
		// FilePath 文件绝对路径，支持文件和文件夹
		FilePath string
	}

	AppGetFileInfoResult struct {
		XMLName xml.Name `xml:"folderInfo"`
		Id string `xml:"id"`
		ParentFolderId string `xml:"parentFolderId"`
		Path string `xml:"path"`
		Name string `xml:"name"`
		CreateDate string `xml:"createDate"`
		LastOpTime string `xml:"lastOpTime"`
		Rev string `xml:"rev"`
		ParentFolderList parentFolderListNode `xml:"parentFolderList"`
	}
	parentFolderListNode struct {
		FolderList []appGetFolderInfoNode `xml:"folder"`
	}
	appGetFolderInfoNode struct {
		Fid string `xml:"fid"`
		Fname string `xml:"fname"`
	}
)

// AppGetFileInfo 根据文件ID或者文件绝对路径获取文件信息，支持文件和文件夹
func (p *PanClient) AppGetFileInfo(param *AppGetFileInfoParam) (*AppGetFileInfoResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/getFolderInfo.action?folderId=%s&folderPath=%s&pathList=1&dt=3&%s",
		API_URL, param.FileId, url.QueryEscape(param.FilePath), apiutil.PcClientInfoSuffixParam())
	httpMethod := "GET"
	dateOfGmt := apiutil.DateOfGmtStr()
	appToken := p.appToken
	headers := map[string]string {
		"Date": dateOfGmt,
		"SessionKey": appToken.SessionKey,
		"Signature": apiutil.SignatureOfHmac(appToken.SessionSecret, appToken.SessionKey, httpMethod, fullUrl.String(), dateOfGmt),
		"X-Request-ID": apiutil.XRequestId(),
	}
	logger.Verboseln("do request url: " + fullUrl.String())
	respBody, err1 := p.client.Fetch(httpMethod, fullUrl.String(), nil, headers)
	if err1 != nil {
		logger.Verboseln("AppGetFileInfo occurs error: ", err1.Error())
		return nil, apierror.NewApiErrorWithError(err1)
	}
	er := &apierror.AppErrorXmlResp{}
	if err := xml.Unmarshal(respBody, er); err == nil {
		if er.Code != "" {
			if er.Code == "FileNotFound" {
				return nil, apierror.NewApiError(apierror.ApiCodeFileNotFoundCode, "文件不存在")
			}
		}
	}
	item := &AppGetFileInfoResult{}
	if err := xml.Unmarshal(respBody, item); err != nil {
		logger.Verboseln("AppGetFileInfo parse response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil

}