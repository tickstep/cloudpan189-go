package upload

import (
	"context"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/internal/file/uploader"
	"github.com/tickstep/cloudpan189-go/library/requester"
	"github.com/tickstep/cloudpan189-go/library/requester/rio"
	"io"
	"net/http"
)

type (
	PCSUpload struct {
		panClient        *cloudpan.PanClient
		targetPath string

		// UploadFileId 上传文件请求ID
		uploadFileId string
		// FileUploadUrl 上传文件数据的URL路径
		fileUploadUrl string
		// FileCommitUrl 上传文件完成后确认路径
		fileCommitUrl string
		// 请求的X-Request-ID
		xRequestId string
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

func NewPCSUpload(pcs *cloudpan.PanClient, targetPath, uploadUrl, commitUrl, uploadFileId, xRequestId string) uploader.MultiUpload {
	return &PCSUpload{
		panClient:        pcs,
		targetPath: targetPath,
		uploadFileId: uploadFileId,
		fileUploadUrl: uploadUrl,
		fileCommitUrl: commitUrl,
		xRequestId: xRequestId,
	}
}

func (pu *PCSUpload) lazyInit() {
	if pu.panClient == nil {
		pu.panClient = &cloudpan.PanClient{}
	}
}

func (pu *PCSUpload) Precreate() (err error) {
	return nil
}

func (pu *PCSUpload) UploadFile(ctx context.Context, partseq int, partOffset int64, partEnd int64, r rio.ReaderLen64) (uploadDone bool, uperr error) {
	pu.lazyInit()

	var respErr *uploader.MultiError
	fileRange := &cloudpan.AppFileRange{
		Start: int(partOffset),
		End: int(partEnd),
	}
	pcsError := pu.panClient.AppUploadFileData(pu.fileUploadUrl, pu.uploadFileId, pu.xRequestId, fileRange,
		func(httpMethod, fullUrl string, headers map[string]string) (resp *http.Response, err error) {
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
		return resp, nil
	})

	if pcsError != nil {
		respErr.Err = pcsError
		return false, respErr
	}

	return true, nil
}

func (pu *PCSUpload) CommitFile() (cerr error) {
	pu.lazyInit()
	_, er := pu.panClient.AppUploadFileCommit(pu.fileCommitUrl, pu.uploadFileId, pu.xRequestId)
	if er != nil {
		return er
	}
	return nil
}
