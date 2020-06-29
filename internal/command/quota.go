package command

type QuotaInfo struct {
	// 已使用个人空间大小
	UsedSize int64
	// 个人空间总大小
	Quota int64
}

func RunGetQuotaInfo() (quotaInfo *QuotaInfo, error error) {
	user, err := GetActivePanClient().GetUserInfo()
	if err != nil {
		return nil, err
	}
	return &QuotaInfo{
		UsedSize: int64(user.UsedSize),
		Quota: int64(user.Quota),
	}, nil
}
