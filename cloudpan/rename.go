package cloudpan

import (
	"encoding/json"
	"fmt"
	"github.com/tickstep/cloudpan189-go/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/library/logger"
	"net/url"
	"strings"
)

func (p *PanClient) Rename(renameFileId, newName string) (bool, *apierror.ApiError) {
	if renameFileId == "" {
		return false, apierror.NewFailedApiError("请指定命名的文件")
	}

	fullUrl := &strings.Builder{}
	fmt.Fprintf(fullUrl, "%s/v2/renameFile.action?fileId=%s&fileName=%s",
		WEB_URL, renameFileId, url.QueryEscape(newName))
	logger.Verboseln("do request url: " + fullUrl.String())
	body, err := p.client.DoGet(fullUrl.String())
	if err != nil {
		logger.Verboseln("Rename failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	comResp := &apierror.ErrorResp{}
	if err := json.Unmarshal(body, comResp); err == nil {
		if comResp.ErrorCode == "FileAlreadyExists" {
			logger.Verboseln("Rename response failed")
			return false, apierror.NewFailedApiError("文件名已存在")
		}
	}

	result := &apierror.SuccessResp{}
	if err := json.Unmarshal(body, result); err != nil {
		logger.Verboseln("Rename response failed")
		return false, apierror.NewApiErrorWithError(err)
	}
	return result.Success, nil
}
