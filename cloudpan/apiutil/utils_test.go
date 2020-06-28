package apiutil

import (
	"github.com/stretchr/testify/assert"
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