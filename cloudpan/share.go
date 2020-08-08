package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"github.com/tickstep/cloudpan189-go/library/text"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type (
	ShareExpiredTime int
	ShareMode int

	PrivateShareResult struct {
		AccessCode string `json:"accessCode"`
		ShortShareUrl string `json:"shortShareUrl"`
	}

	PublicShareResult struct {
		ShareId int `json:"shareId"`
		ShortShareUrl string `json:"shortShareUrl"`
	}

	AccessCount struct {
		CopyCount int `json:"copyCount"`
		DownloadCount int `json:"downloadCount"`
		PreviewCount int `json:"previewCount"`
	}

	ShareItem struct {
		// AccessCode 提取码，私密分享才有
		AccessCode string `json:"accessCode"`
		// AccessURL 分享链接
		AccessURL string `json:"accessURL"`
		// AccessCount 分享被查看下载次数
		AccessCount AccessCount `json:"accessCount"`
		// DownloadUrl 下载路径，文件才会有
		DownloadUrl string `json:"downloadUrl"`
		// DownloadUrl 下载路径，文件才会有
		LongDownloadUrl string `json:"longDownloadUrl"`
		// FileId 文件ID
		FileId string `json:"fileId"`
		// FileIdDigest 文件指纹
		FileIdDigest string `json:"fileIdDigest"`
		// FileName 文件名
		FileName string `json:"fileName"`
		// FilePath 路径
		FilePath string `json:"filePath"`
		// FileSize 文件大小，文件夹为0
		FileSize int64 `json:"fileSize"`
		// IconURL 缩略图路径???
		IconURL string `json:"iconURL"`
		// IsFolder 是否是文件夹
		IsFolder bool `json:"isFolder"`
		// MediaType 文件类别
		MediaType MediaType `json:"mediaType"`
		NeedAccessCode int `json:"needAccessCode"`
		// NickName 分享者账号昵称
		NickName string `json:"nickName"`
		// ReviewStatus 审查状态，1-正常
		ReviewStatus int `json:"reviewStatus"`
		// ShareDate 分享日期
		ShareDate string `json:"shareDate"`
		// ShareId 分享项目ID，唯一标识该分享项
		ShareId int64 `json:"shareId"`
		// ShareMode 分享模式，1-私密，2-公开
		ShareMode ShareMode `json:"shareMode"`
		// ShareTime 分享时间
		ShareTime string `json:"shareTime"`
		// ShareType 分享类别，默认都是1
		ShareType int `json:"shareType"`
		// ShortShareUrl 分享的访问路径，和 AccessURL 一致
		ShortShareUrl string `json:"shortShareUrl"`
	}

	ShareItemList []*ShareItem

	// ShareListResult 获取分享项目列表响应体
	ShareListResult struct {
		Data ShareItemList `json:"data"`
		PageNum int `json:"pageNum"`
		PageSize int `json:"pageSize"`
		RecordCount int `json:"recordCount"`
	}

	ShareListParam struct {
		ShareType int `json:"shareType"`
		PageNum int `json:"pageNum"`
		PageSize int `json:"pageSize"`
	}

	errResp struct {
		ErrorVO apierror.ErrorResp `json:"errorVO"`
	}

	ShareListDirResult struct {
		AccessCount AccessCount `json:"accessCount"`
		Data FileList `json:"data"`
		Digest string `json:"digest"`
		ExpireTime int `json:"expireTime"`
		ExpireType int `json:"expireType"`
		PageNum int `json:"pageNum"`
		PageSize int `json:"pageSize"`
		RecordCount int `json:"recordCount"`
		ShareDate string `json:"shareDate"`
	}

)

const (
	// 1天期限
	ShareExpiredTime1Day ShareExpiredTime = 1
	// 7天期限
	ShareExpiredTime7Day ShareExpiredTime = 7
	// 永久期限
	ShareExpiredTimeForever ShareExpiredTime = 2099

	// ShareModePrivate 私密分享
	ShareModePrivate ShareMode = 1
	// ShareModePublic 公开分享
	ShareModePublic ShareMode = 2
)

func (p *PanClient) SharePrivate(fileId string, expiredTime ShareExpiredTime) (*PrivateShareResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/privateLinkShare.action?fileId=%s&expireTime=%d&withAccessCode=1",
		WEB_URL, fileId, expiredTime)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	logger.Verboseln("response body: " + string(body))
	if err != nil {
		logger.Verboseln("SharePrivate failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	errResp := &errResp{}
	if err := json.Unmarshal(body, errResp); err == nil {
		if errResp.ErrorVO.ErrorCode != "" {
			logger.Verboseln("SharePrivate response failed")
			if errResp.ErrorVO.ErrorCode == "ShareCreateOverload" {
				return nil, apierror.NewFailedApiError("您分享的次数已达上限，请明天再来吧")
			}
			return nil, apierror.NewApiErrorWithError(err)
		}
	}

	item := &PrivateShareResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("SharePrivate response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

func (p *PanClient) SharePublic(fileId string, expiredTime ShareExpiredTime) (*PublicShareResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/createOutLinkShare.action?fileId=%s&expireTime=%d&withAccessCode=1",
		WEB_URL, fileId, expiredTime)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("SharePublic failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &PublicShareResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("SharePublic response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

func NewShareListParam() *ShareListParam {
	return &ShareListParam{
		ShareType: 1,
		PageNum: 1,
		PageSize: 60,
	}
}
func (p *PanClient) ShareList(param *ShareListParam) (*ShareListResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/listShares.action?shareType=%d&pageNum=%d&pageSize=%d",
		WEB_URL, param.ShareType, param.PageNum, param.PageSize)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("ShareList failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &ShareListResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("ShareList response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	// normalize
	for _, s := range item.Data {
		s.AccessURL = "https:" + s.AccessURL
		s.ShortShareUrl = "https:" + s.ShortShareUrl
	}
	return item, nil
}

func (p *PanClient) ShareCancel(shareIdList []int64) (bool, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	shareIds := ""
	for _, id := range shareIdList {
		shareIds += strconv.Itoa(int(id)) + ","
	}
	if strings.LastIndex(shareIds, ",") == (len(shareIds) - 1) {
		shareIds = text.Substr(shareIds, 0, len(shareIds) - 1)
	}

	fmt.Fprintf(fullUrl, "%s/v2/cancelShare.action?shareIdList=%s&ancelType=1",
		WEB_URL, url.QueryEscape(shareIds))
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("ShareCancel failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	comResp := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, comResp); err == nil {
		if comResp.ErrorCode != "" {
			logger.Verboseln("ShareCancel response failed")
			return false, apierror.NewFailedApiError("取消分享失败，请稍后重试")
		}
	}
	item := &apierror.SuccessResp{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("ShareCancel response failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	return item.Success, nil
}

func (p *PanClient) ShareGetIdByUrl(accessUrl string) (int64, string, *apierror.ApiError) {
	if strings.Index(accessUrl, WEB_URL) < 0 {
		return 0, "", apierror.NewFailedApiError("URL错误")
	}
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s", accessUrl)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("ShareGetIdByUrl failed")
		return 0, "", apierror.NewApiErrorWithError(err)
	}

	htmlText := string(body)

	re, _ := regexp.Compile("_shareId = '(.+?)'")
	shareIdStr := re.FindStringSubmatch(htmlText)[1]
	shareId := 0
	if shareIdStr != "" {
		shareId,_ = strconv.Atoi(shareIdStr)
	}

	re, _ = regexp.Compile("_verifyCode = '(.+?)'")
	verifyCodeStr := re.FindStringSubmatch(htmlText)[1]

	return int64(shareId), verifyCodeStr, nil
}

func (p *PanClient) ShareListDirDetail(accessUrl string, accessCode string) (int64, *ShareListDirResult, *apierror.ApiError) {
	shareId, verifyCode, apierr := p.ShareGetIdByUrl(accessUrl)
	if apierr != nil {
		return 0, nil, apierr
	}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/listShareDir.action?shareId=%d&accessCode=%s&verifyCode=%s&orderBy=1&order=ASC&pageNum=1&pageSize=60 ",
		WEB_URL, shareId, accessCode, verifyCode)

	header := map[string]string {
		"Referer": accessUrl,
	}

	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := client.Fetch("GET", fullUrl.String(), nil, header)
	if err != nil {
		logger.Verboseln("ShareListDirDetail failed")
		return 0, nil, apierror.NewApiErrorWithError(err)
	}

	item := &ShareListDirResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("ShareListDirDetail response failed")
		return 0, nil, apierror.NewApiErrorWithError(err)
	}
	return shareId, item, nil
}

// ShareSave 转存分享到对应的文件夹
func (p *PanClient) ShareSave(accessUrl string, accessCode string, savePanDirId string) (bool, *apierror.ApiError) {
	shareId, shareListDirResult, apierror := p.ShareListDirDetail(accessUrl, accessCode)
	if apierror != nil {
		return false, apierror
	}

	taskReqParam := &BatchTaskParam{
		TypeFlag: BatchTaskTypeShareSave,
		TaskInfos: makeBatchTaskInfoListForShareSave(shareListDirResult.Data),
		TargetFolderId: savePanDirId,
		ShareId: shareId,
	}
	taskId, apierror1 := p.CreateBatchTask(taskReqParam)
	logger.Verboseln("share save taskid: ", taskId)
	return taskId != "", apierror1
}

func makeBatchTaskInfoListForShareSave(opFileList FileList) (infoList BatchTaskInfoList) {
	for _, fe := range opFileList {
		isFolder := 0
		if fe.IsFolder {
			isFolder = 1
		}
		infoItem := &BatchTaskInfo{
			FileId: fe.FileId,
			FileName: fe.FileName,
			IsFolder: isFolder,
			SrcParentId: fe.ParentId,
		}
		infoList = append(infoList, infoItem)
	}
	return
}