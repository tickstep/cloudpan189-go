package upload

import (
	"context"
	"github.com/tickstep/cloudpan189-go/cloudpan"
	"github.com/tickstep/cloudpan189-go/internal/config"
	"github.com/tickstep/cloudpan189-go/library/requester"
	"github.com/tickstep/cloudpan189-go/library/requester/rio"
	"github.com/tickstep/cloudpan189-go/internal/file/uploader"
	"io"
	"net/http"
)

type (
	PCSUpload struct {
		panClient        *cloudpan.PanClient
		targetPath string
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

func NewPCSUpload(pcs *cloudpan.PanClient, targetPath string) uploader.MultiUpload {
	return &PCSUpload{
		panClient:        pcs,
		targetPath: targetPath,
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

func (pu *PCSUpload) TmpFile(ctx context.Context, partseq int, partOffset int64, r rio.ReaderLen64) (uploadDone bool, uperr error) {
	pu.lazyInit()

	var respErr *uploader.MultiError
	pcsError := pu.panClient.AppUploadFileData(,
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
				case 400, 401, 403, 413:
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
	})

	if respErr != nil {
		respErr.Err = pcsError
		return false, respErr
	}

	return true, pcsError
}

func (pu *PCSUpload) CreateSuperFile(checksumList ...string) (err error) {
	pu.lazyInit()
	return nil
}
