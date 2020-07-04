package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"net/url"
	"strings"
)

type (
	MkdirResult struct {
		// fileId 文件ID
		FileId string `json:"fileId"`
		// isNew 是否创建成功。true为成功，false或者没有返回则为失败，失败原因基本是已存在该文件夹
		IsNew bool `json:"isNew"`
	}
)

func (p *PanClient) Mkdir(parentFileId, dirName string) (*MkdirResult, *apierror.ApiError) {
	if parentFileId == "" {
		// 默认根目录
		parentFileId = "-11"
	}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/createFolder.action?parentId=%s&fileName=%s",
		WEB_URL, parentFileId, url.QueryEscape(dirName))
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("mkdir failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	item := &MkdirResult{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("mkdir response failed")
		return nil, apierror.NewApiErrorWithError(err)
	}
	if !item.IsNew {
		return item, apierror.NewFailedApiError("文件夹已存在: " + dirName)
	}
	return item, nil
}
