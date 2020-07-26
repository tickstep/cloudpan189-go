package ids

import (
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"strings"
)

const (
	DefaultUniqueId = "11884a11d126t51b74a1f99b91da91814c56a588fe6bf11f9224600e3a400fcc"
)

func GetUniqueId(appId string, size int) string {
	uid := DefaultUniqueId

	var (
		id = ""
		err = fmt.Errorf("")
	)

	if appId == "" {
		id, err = machineid.ID()
		if err == nil {
			uid = strings.ReplaceAll(id, "-", "")
		}
	} else {
		id, err = machineid.ProtectedID("cloudpan189-go")
		if err == nil {
			uid = id
		}
	}

	if size > 0 && size < len(uid) {
		return uid[:size]
	}
	return uid
}
