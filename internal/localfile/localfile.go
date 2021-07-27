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
package localfile

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/tickstep/cloudpan189-api/cloudpan"
	"hash/crc32"
	"io"
	"os"

	"github.com/tickstep/library-go/cachepool"
	"github.com/tickstep/library-go/converter"
)

const (
	// DefaultBufSize 默认的bufSize
	DefaultBufSize = int(256 * converter.KB)
)

const (
	// CHECKSUM_MD5 获取文件的 md5 值
	CHECKSUM_MD5 int = 1 << iota

	// CHECKSUM_CRC32 获取文件的 crc32 值
	CHECKSUM_CRC32
)

type (
	// LocalFileMeta 本地文件元信息
	LocalFileMeta struct {
		Path    string `json:"path,omitempty"`   // 本地路径
		Length  int64  `json:"length,omitempty"` // 文件大小
		MD5     string `json:"md5,omitempty"`    // 文件的 md5
		CRC32   uint32 `json:"crc32,omitempty"`  // 文件的 crc32
		ModTime int64  `json:"modtime"`          // 修改日期

		// ParentFolderId 存储云盘的目录ID
		ParentFolderId string `json:"parent_folder_id,omitempty"`
		// UploadFileId 上传文件请求ID
		UploadFileId string `json:"upload_file_id,omitempty"`
		// FileUploadUrl 上传文件数据的URL路径
		FileUploadUrl string `json:"file_upload_url,omitempty"`
		// FileCommitUrl 上传文件完成后确认路径
		FileCommitUrl string `json:"file_commit_url,omitempty"`
		// FileDataExists 文件是否已存在云盘中，0-未存在，1-已存在
		FileDataExists int `json:"file_data_exists,omitempty"`
		// 请求的X-Request-ID
		XRequestId string `json:"x_request_id,omitempty"`
	}

	// LocalFileEntity 校验本地文件
	LocalFileEntity struct {
		LocalFileMeta
		bufSize int
		buf     []byte
		file    *os.File // 文件
	}
)

func NewLocalFileEntity(localPath string) *LocalFileEntity {
	return NewLocalFileEntityWithBufSize(localPath, DefaultBufSize)
}

func NewLocalFileEntityWithBufSize(localPath string, bufSize int) *LocalFileEntity {
	return &LocalFileEntity{
		LocalFileMeta: LocalFileMeta{
			Path: localPath,
		},
		bufSize: bufSize,
	}
}

// OpenPath 检查文件状态并获取文件的大小 (Length)
func (lfc *LocalFileEntity) OpenPath() error {
	if lfc.file != nil {
		lfc.file.Close()
	}

	var err error
	lfc.file, err = os.Open(lfc.Path)
	if err != nil {
		return err
	}

	info, err := lfc.file.Stat()
	if err != nil {
		return err
	}

	lfc.Length = info.Size()
	lfc.ModTime = info.ModTime().Unix()
	return nil
}

// GetFile 获取文件
func (lfc *LocalFileEntity) GetFile() *os.File {
	return lfc.file
}

// Close 关闭文件
func (lfc *LocalFileEntity) Close() error {
	if lfc.file == nil {
		return ErrFileIsNil
	}

	return lfc.file.Close()
}

func (lfc *LocalFileEntity) initBuf() {
	if lfc.buf == nil {
		lfc.buf = cachepool.RawMallocByteSlice(lfc.bufSize)
	}
}

func (lfc *LocalFileEntity) writeChecksum(data []byte, wus ...*ChecksumWriteUnit) (err error) {
	doneCount := 0
	for _, wu := range wus {
		_, err := wu.Write(data)
		switch err {
		case ErrChecksumWriteStop:
			doneCount++
			continue
		case nil:
		default:
			return err
		}
	}
	if doneCount == len(wus) {
		return ErrChecksumWriteAllStop
	}
	return nil
}

func (lfc *LocalFileEntity) repeatRead(wus ...*ChecksumWriteUnit) (err error) {
	if lfc.file == nil {
		return ErrFileIsNil
	}

	lfc.initBuf()

	defer func() {
		_, err = lfc.file.Seek(0, os.SEEK_SET) // 恢复文件指针
		if err != nil {
			return
		}
	}()

	// 读文件
	var (
		n int
	)
read:
	for {
		n, err = lfc.file.Read(lfc.buf)
		switch err {
		case io.EOF:
			err = lfc.writeChecksum(lfc.buf[:n], wus...)
			break read
		case nil:
			err = lfc.writeChecksum(lfc.buf[:n], wus...)
		default:
			return
		}
	}
	switch err {
	case ErrChecksumWriteAllStop: // 全部结束
		err = nil
	}
	return
}

func (lfc *LocalFileEntity) createChecksumWriteUnit(cw ChecksumWriter, isAll bool, getSumFunc func(sum interface{})) (wu *ChecksumWriteUnit, deferFunc func(err error)) {
	wu = &ChecksumWriteUnit{
		ChecksumWriter: cw,
		End:            lfc.LocalFileMeta.Length,
		OnlySliceSum:   !isAll,
	}

	return wu, func(err error) {
		if err != nil {
			return
		}
		getSumFunc(wu.Sum)
	}
}

// Sum 计算文件摘要值
func (lfc *LocalFileEntity) Sum(checkSumFlag int) (err error) {
	lfc.fix()
	wus := make([]*ChecksumWriteUnit, 0, 2)
	if (checkSumFlag & (CHECKSUM_MD5)) != 0 {
		md5w := md5.New()
		wu, d := lfc.createChecksumWriteUnit(
			NewHashChecksumWriter(md5w),
			(checkSumFlag&CHECKSUM_MD5) != 0,
			func(sum interface{}) {
				if sum != nil {
					lfc.MD5 = hex.EncodeToString(sum.([]byte))
				}

				// zero size file
				if lfc.Length == 0 {
					lfc.MD5 = cloudpan.DefaultEmptyFileMd5
				}
			},
		)

		wus = append(wus, wu)
		defer d(err)
	}
	if (checkSumFlag & CHECKSUM_CRC32) != 0 {
		crc32w := crc32.NewIEEE()
		wu, d := lfc.createChecksumWriteUnit(
			NewHash32ChecksumWriter(crc32w),
			true,
			func(sum interface{}) {
				if sum != nil {
					lfc.CRC32 = sum.(uint32)
				}
			},
		)

		wus = append(wus, wu)
		defer d(err)
	}

	err = lfc.repeatRead(wus...)
	return
}

func (lfc *LocalFileEntity) fix() {
	if lfc.bufSize < DefaultBufSize {
		lfc.bufSize = DefaultBufSize
	}
}
