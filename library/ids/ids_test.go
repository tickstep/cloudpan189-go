package ids

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetUniqueId(t *testing.T) {
	fmt.Println(GetUniqueId("", 0))
	assert.Equal(t, 32, len(GetUniqueId("", 0)))
}

func TestGetUniqueIdWithAppId(t *testing.T) {
	fmt.Println(GetUniqueId("cloudpan189", 0))
	assert.Equal(t, 64, len(GetUniqueId("cloudpan189", 0)))
}
