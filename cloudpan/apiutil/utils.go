package apiutil

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"math/rand"
	"net/http"
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

	FileNameSpecialChars = "\\/:*?\"<>|"
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

// SignatureOfMd5 MD5签名
func SignatureOfMd5(params map[string]string) string {
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

// SignatureOfHmac HMAC签名
func SignatureOfHmac(secretKey, sessionKey, operate, url, dateOfGmt string) string {
	requestUri := strings.ReplaceAll(strings.Split(url, "?")[0], "https://api.cloud.189.cn", "")
	plainStr := &strings.Builder{}
	fmt.Fprintf(plainStr, "SessionKey=%s&Operate=%s&RequestURI=%s&Date=%s",
		sessionKey, operate, requestUri, dateOfGmt)

	key := []byte(secretKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(plainStr.String()))
	return strings.ToUpper(hex.EncodeToString(mac.Sum(nil)))
}

func Rand() string {
	randStr := &strings.Builder{}
	fmt.Fprintf(randStr, "%d_%d", rand.Int63n(1e5), rand.Int63n(1e10))
	return randStr.String()
}

// PcClientInfoSuffixParam PC客户端URL请求后缀信息
func PcClientInfoSuffixParam() string {
	return "clientType=TELEPC&version=6.2&channelId=web_cloud.189.cn&rand=" + Rand()
}

func DateOfGmtStr() string {
	return time.Now().UTC().Format(http.TimeFormat)
}

func XRequestId() string {
	u4 := uuid.NewV4()
	return strings.ToUpper(u4.String())
}

func Uuid() string {
	u4 := uuid.NewV4()
	return u4.String()
}

// CheckFileNameValid 检测文件名是否有效，包含特殊字符则无效
func CheckFileNameValid(name string) bool {
	if name == "" {
		return true
	}
	return !strings.ContainsAny(name, FileNameSpecialChars)
}