// Copyright (c) 2020 tickstep.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
