package config

import (
	"fmt"
	"testing"
)

func TestEncryptString(t *testing.T) {
	fmt.Println(EncryptString("131687xxxxx@189.cn"))
}

func TestDecryptString(t *testing.T) {
	fmt.Println(DecryptString("75b3c8d21607440c0e8a70f4a4861c8669774cc69c70ce2a2c8acb815b6d5d3b"))
}
