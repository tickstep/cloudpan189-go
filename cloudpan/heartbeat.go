package cloudpan

import (
	"encoding/json"
	"github.com/tickstep/cloudpan189-go/library/logger"
)

type heartBeatResp struct {
	Success bool `json:"success"`
}

// Heartbeat WEB端心跳包，周期默认1分钟
func (p *PanClient) Heartbeat() bool  {
	url := WEB_URL + "/heartbeat.action"
	body, err := p.client.DoGet(url)
	if err != nil {
		logger.Verboseln("heartbeat failed")
		return false
	}
	item := &heartBeatResp{}
	if err := json.Unmarshal(body, item); err != nil {
		logger.Verboseln("heartbeat response failed")
		return false
	}
	return item.Success
}
