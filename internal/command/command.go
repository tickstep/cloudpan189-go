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

import (
	"errors"
	"strconv"

	"github.com/urfave/cli"

	"github.com/tickstep/cloudpan189-api/cloudpan"
	"github.com/tickstep/cloudpan189-go/internal/config"
)

var ErrBadArgs = errors.New("参数错误")
var ErrNotLogined = errors.New("未登录账号")

func GetActivePanClient() *cloudpan.PanClient {
	return config.Config.ActiveUser().PanClient()
}

func GetActiveUser() *config.PanUser {
	return config.Config.ActiveUser()
}

func parseFamilyId(c *cli.Context) int64 {
	familyId := config.Config.ActiveUser().ActiveFamilyId
	if c.IsSet("familyId") {
		fid, errfi := strconv.ParseInt(c.String("familyId"), 10, 64)
		if errfi != nil {
			familyId = 0
		} else {
			familyId = fid
		}
	}
	return familyId
}
