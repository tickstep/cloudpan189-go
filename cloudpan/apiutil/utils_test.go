package apiutil

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestB64toHex(t *testing.T) {
	bs64 := "pi7yXRON3QtNuvDzTJ7jpkaA+tGYNTwMIKeecDQbKLCjeXHxN3HBmZJbX3BFlGG81hsfliewDuX/clVKHYHMQxz9L1VVVHB4Oyi9504cn9Tlc4qT2r9whcjZADcqPLms3m3+WWCUAGLVsKj7Flqs2OfuoQSuFaqhZ31t69xgbeQ="
	hex := B64toHex(bs64)
	//fmt.Println(hex)
	assert.Equal(t,
		"a62ef25d138ddd0b4dbaf0f34c9ee3a64680fad198353c0c20a79e70341b28b0a37971f13771c199925b5f70459461bcd61b1f9627b00ee5ff72554a1d81cc431cfd2f55555470783b28bde74e1c9fd4e5738a93dabf7085c8d900372a3cb9acde6dfe5960940062d5b0a8fb165aacd8e7eea104ae15aaa1677d6debdc606de4",
		hex)
}

func TestNoCache(t *testing.T) {
	noCache := noCache()
	fmt.Println(noCache)
	assert.Equal(t, len("0.") + 17, len(noCache))
}

func TestTimestampe(t *testing.T) {
	r := Timestamp()
	fmt.Println(r)
	assert.Equal(t, 13, len(strconv.Itoa(r)))
}

func TestSignatureOfMd5(t *testing.T) {
	params := map[string]string {
		"Timestamp": "1593905856153",
		"sessionKey": "c99af8b0-cee7-46b8-8fe8-f11ff69417f8",
		"AppKey": "601102120",
	}
	r := SignatureOfMd5(params)
	fmt.Println(r)
	assert.Equal(t, "8f0c6eb9048c087b9f2b6e190afc1140", r)
}

func TestSignatureOfHmac(t *testing.T) {
	r := SignatureOfHmac(
		"01DB3448B69173020D01CCE8BD6EE641",
		"a10c3904-b5d8-403e-ae50-aa55b0e1bfd1",
		"GET",
		"https://api.cloud.189.cn/getUserInfo.action?clientType=TELEPC&version=6.2.3.0&channelId=web_cloud.189.cn&rand=22722_1943588875",
		"Mon, 06 Jul 2020 14:23:47 GMT")
	fmt.Println(r)
	assert.Equal(t, "c37b31af82eb77cc56a20a3d376da98de96b7749", r)
}

func TestRand(t *testing.T) {
	r := Rand()
	fmt.Println(r)
	assert.Equal(t, 16, len(r))
}

func TestDateOfGmtStr(t *testing.T) {
	r := DateOfGmtStr()
	fmt.Println(r)
}