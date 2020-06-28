package crypto

import (
	"bytes"
	"encoding/base64"
	"io/ioutil"
)

// Base64Encode base64加密
func Base64Encode(raw []byte) []byte {
	var encoded bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &encoded)
	encoder.Write(raw)
	encoder.Close()
	return encoded.Bytes()
}

// Base64EncodeStr base64字符串加密
func Base64EncodeStr(raw string) string {
	return string(Base64Encode([]byte(raw)))
}

// Base64Decode base64解密
func Base64Decode(raw []byte) []byte {
	var buf bytes.Buffer
	buf.Write(raw)
	decoder := base64.NewDecoder(base64.StdEncoding, &buf)
	decoded, _ := ioutil.ReadAll(decoder)
	return decoded
}

// Base64DecodeStr base64字符串解密
func Base64DecodeStr(bs64str string) string {
	return string(Base64Decode([]byte(bs64str)))
}