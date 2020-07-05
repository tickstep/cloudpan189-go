package apiutil

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

const (
	RsaPublicKey = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDY7mpaUysvgQkbp0iIn2ezoUyh
i1zPFn0HCXloLFWT7uoNkqtrphpQ/63LEcPz1VYzmDuDIf3iGxQKzeoHTiVMSmW6
FlhDeqVOG094hFJvZeK4OzA6HVwzwnEW5vIZ7d+u61RV1bsFxmB68+8JXs3ycGcE
4anY+YzZJcyOcEGKVQIDAQAB
-----END PUBLIC KEY-----`

	b64map = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	bi_rm = "0123456789abcdefghijklmnopqrstuvwxyz"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func int2char(i int) (r byte) {
	return bi_rm[i]
}

// B64toHex 将base64字符串转换成HEX十六进制字符串
func B64toHex(b64str string) (hexstr string) {
	sb := strings.Builder{}
	e := 0
	c := 0
	for _,r := range b64str {
		if r != '=' {
			v := strings.Index(b64map, string(r))
			if 0 == e {
				e = 1
				sb.WriteByte(int2char(v >> 2))
				c = 3 & v
			} else if 1 == e {
				e = 2
				sb.WriteByte(int2char(c << 2 | v >> 4))
				c = 15 & v
			} else if 2 == e {
				e = 3
				sb.WriteByte(int2char(c))
				sb.WriteByte(int2char(v >> 2))
				c = 3 & v
			} else {
				e = 0
				sb.WriteByte(int2char(c << 2 | v >> 4))
				sb.WriteByte(int2char(15 & v))
			}
		}
	}
	if e == 1 {
		sb.WriteByte(int2char(c << 2))
	}
	return sb.String()
}

func noCache() string {
	noCache := &strings.Builder{}
	fmt.Fprintf(noCache, "0.%d", rand.Int63n(1e17))
	return noCache.String()
}

func Timestamp() int {
	// millisecond
	return int(time.Now().UTC().UnixNano() / 1e6)
}

// Signature MD5签名
func Signature(params map[string]string) string {
	keys := []string{}
	for k, v := range params {
		keys = append(keys, k + "=" + v)
	}

	// sort
	sort.Strings(keys)

	signStr := strings.Join(keys, "&")

	h := md5.New()
	h.Write([]byte(signStr))
	return hex.EncodeToString(h.Sum(nil))
}