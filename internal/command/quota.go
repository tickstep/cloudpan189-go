package command

import (
	"github.com/tickstep/cloudpan189-go/cloudpan/apiweb"
	"github.com/tickstep/cloudpan189-go/internal/config"
)

type QuotaInfo struct {
	// 已使用个人空间大小
	UsedSize int64
	// 个人空间总大小
	Quota int64
}

func RunGetQuotaInfo() (quotaInfo *QuotaInfo, error error) {
	user, err := apiweb.GetUserInfo(config.Config.ActiveUser().PanClient())
	if err != nil {
		return nil, err
	}
	return &QuotaInfo{
		UsedSize: int64(user.UsedSize),
		Quota: int64(user.Quota),
	}, nil
}
