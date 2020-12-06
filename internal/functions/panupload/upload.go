// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package panupload

import (
	"context"
	"io"
	"net/http"

	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-api/cloudpan/apierror"
	"github.com/tickstep/cloudpan189-go/internal/file/uploader"
	"github.com/tickstep/library-go/requester"
	"github.com/tickstep/library-go/requester/rio"
)

type (
	PanUpload struct {
		panClient  *cloudpan.PanClient
		targetPath string
		familyId   int64

		// UploadFileId 上传文件请求ID
		uploadFileId string
		// FileUploadUrl 上传文件数据的URL路径
		fileUploadUrl string
		// FileCommitUrl 上传文件完成后确认路径
		fileCommitUrl string
		// 请求的X-Request-ID
		xRequestId string
	}

	UploadedFileMeta struct {
		IsFolder     bool   `json:"isFolder,omitempty"` // 是否目录
		Path         string `json:"-"`                  // 本地路径，不记录到数据库
		MD5          string `json:"md5,omitempty"`      // 文件的 md5
		FileID       string `json:"id,omitempty"`       //文件、目录ID
		ParentId     string `json:"parentId,omitempty"` //父文件夹ID
		Rev          string `json:"rev,omitempty"`      //文件版本
		Size         int64  `json:"length,omitempty"`   // 文件大小
		ModTime      int64  `json:"modtime,omitempty"`  // 修改日期
		LastSyncTime int64  `json:"synctime,omitempty"` //最后同步时间
	}

	EmptyReaderLen64 struct {
	}
)

func (e EmptyReaderLen64) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (e EmptyReaderLen64) Len() int64 {
	return 0
}

func NewPanUpload(panClient *cloudpan.PanClient, targetPath, uploadUrl, commitUrl, uploadFileId, xRequestId string, familyId int64) uploader.MultiUpload {
	return &PanUpload{
		panClient:     panClient,
		targetPath:    targetPath,
		familyId:      familyId,
		uploadFileId:  uploadFileId,
		fileUploadUrl: uploadUrl,
		fileCommitUrl: commitUrl,
		xRequestId:    xRequestId,
	}
}

func (pu *PanUpload) lazyInit() {
	if pu.panClient == nil {
		pu.panClient = &cloudpan.PanClient{}
	}
}

func (pu *PanUpload) Precreate() (err error) {
	return nil
}

func (pu *PanUpload) UploadFile(ctx context.Context, partseq int, partOffset int64, partEnd int64, r rio.ReaderLen64) (uploadDone bool, uperr error) {
	pu.lazyInit()

	var respErr *uploader.MultiError
	fileRange := &cloudpan.AppFileUploadRange{
		Offset: partOffset,
		Len:    partEnd - partOffset,
	}
	var apiError *apierror.ApiError
	uploadFunc := func(httpMethod, fullUrl string, headers map[string]string) (resp *http.Response, err error) {
		client := requester.NewHTTPClient()
		client.SetTimeout(0)

		doneChan := make(chan struct{}, 1)
		go func() {
			resp, err = client.Req(httpMethod, fullUrl, r, headers)
			doneChan <- struct{}{}

			if resp != nil {
				// 不可恢复的错误
				switch resp.StatusCode {
				case 400, 401, 403, 413, 600:
					respErr = &uploader.MultiError{
						Terminated: true,
					}
				}
			}
		}()
		select {
		case <-ctx.Done(): // 取消
			// 返回, 让那边关闭连接
			return resp, ctx.Err()
		case <-doneChan:
			// return
		}
		return
	}
	if pu.familyId > 0 {
		apiError = pu.panClient.AppFamilyUploadFileData(pu.familyId, pu.fileUploadUrl, pu.uploadFileId, pu.xRequestId, fileRange, uploadFunc)
	} else {
		apiError = pu.panClient.AppUploadFileData(pu.fileUploadUrl, pu.uploadFileId, pu.xRequestId, fileRange, uploadFunc)
	}

	if respErr != nil {
		return false, respErr
	}

	if apiError != nil {
		return false, apiError
	}

	return true, nil
}

func (pu *PanUpload) CommitFile() (cerr error) {
	pu.lazyInit()
	var er *apierror.ApiError
	if pu.familyId > 0 {
		_, er = pu.panClient.AppFamilyUploadFileCommit(pu.familyId, pu.fileCommitUrl, pu.uploadFileId, pu.xRequestId)
	} else {
		_, er = pu.panClient.AppUploadFileCommit(pu.fileCommitUrl, pu.uploadFileId, pu.xRequestId)
	}
	if er != nil {
		return er
	}
	return nil
}
