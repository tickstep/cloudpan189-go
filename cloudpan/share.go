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