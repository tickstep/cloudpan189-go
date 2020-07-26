package crypto

import (
	"encoding/hex"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncrypt(t *testing.T) {
	d := []byte("131687xxxxx@189.cn")
	key := []byte("16384aeed126e11b74a1f99b91da91814c56a588fe6bf11f92946c0e3a400f5f")[:16]
	fmt.Println(len(key))
	r, e := EncryptAES(d, key)
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Printf("%+v", hex.EncodeToString(r))
	assert.Equal(t, "f13cae8f25083526db1b14a148d8822ea94e65ff2785b99f9e69a835bebdf6ee", hex.EncodeToString(r))
}

func TestDecrypt(t *testing.T) {
	d, _  := hex.DecodeString("f13cae8f25083526db1b14a148d8822ea94e65ff2785b99f9e69a835bebdf6ee")
	key := []byte("16384aeed126e11b74a1f99b91da91814c56a588fe6bf11f92946c0e3a400f5f")[:16]
	r, e := DecryptAES(d, key)
	if e != nil {
		fmt.Println(e)
		return
	}
	fmt.Printf("%s", string(r))
	assert.Equal(t, "131687xxxxx@189.cn", string(r))
}
