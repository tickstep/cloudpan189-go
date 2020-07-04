package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"strings"
)

type (
	ShareExpiredTime int

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
		ShareId int `json:"shareId"`
		// ShareMode 分享模式，1-私密，2-公开
		ShareMode int `json:"shareMode"`
		// ShareTime 分享时间
		ShareTime string `json:"shareTime"`
		// ShareType 分享类别，默认都是1
		ShareType int `json:"shareType"`
		// ShortShareUrl 分享的访问路径，和 AccessURL 一致
		ShortShareUrl string `json:"shortShareUrl"`
	}

	ShareItemList []*ShareItem

	// ListShareItemResult 获取分享项目列表响应体
	ListShareItemResult struct {
		Data ShareItemList `json:"data"`
		PageNum int `json:"pageNum"`
		PageSize int `json:"pageSize"`
		RecordCount int `json:"recordCount"`
	}

	ListShareItemParam struct {
		ShareType int `json:"shareType"`
		PageNum int `json:"pageNum"`
		PageSize int `json:"pageSize"`
	}

)

const (
	// 1天期限
	ShareExpiredTime1Day ShareExpiredTime = 1
	// 7天期限
	ShareExpiredTime7Day ShareExpiredTime = 7
	// 永久期限
	ShareExpiredTimeForever ShareExpiredTime = 2099
)

func (p *PanClient) PrivateShare(fileId string, expiredTime ShareExpiredTime) (*PrivateShareResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/privateLinkShare.action?fileId=%s&expireTime=%d&withAccessCode=1",
		WEB_URL, fileId, expiredTime)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("PrivateShare failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &PrivateShareResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("PrivateShare response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

func (p *PanClient) PublicShare(fileId string, expiredTime ShareExpiredTime) (*PublicShareResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/createOutLinkShare.action?fileId=%s&expireTime=%d&withAccessCode=1",
		WEB_URL, fileId, expiredTime)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("PublicShare failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &PublicShareResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("PublicShare response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}

func NewListShareItemParam() *ListShareItemParam {
	return &ListShareItemParam{
		ShareType: 1,
		PageNum: 1,
		PageSize: 60,
	}
}
func (p *PanClient) ListShare(param *ListShareItemParam) (*ListShareItemResult, *apierror.ApiError) {
	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/listShares.action?shareType=%d&pageNum=%d&pageSize=%d",
		WEB_URL, param.ShareType, param.PageNum, param.PageSize)
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("ListShare failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &ListShareItemResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("ListShare response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	return item, nil
}