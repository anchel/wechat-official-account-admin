package replyservice

import (
	"encoding/json"

	weixinservice "github.com/anchel/wechat-official-account-admin/services/weixin-service"
)

func ParseAutoReplyData(str string) (*weixinservice.AutoReplyData, error) {
	var ret weixinservice.AutoReplyData
	ret.MsgList = make([]*weixinservice.AutoReplyMessage, 0)
	if str == "" {
		return &ret, nil
	}
	err := json.Unmarshal([]byte(str), &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
