package hash

import (
	"crypto/md5"
	"encoding/hex"
)

func Md5OfBytes(data []byte) string {
	h := md5.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}
