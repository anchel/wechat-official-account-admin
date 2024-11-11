package wxapi

import (
	"context"

	"github.com/go-resty/resty/v2"
)

// 发送客服消息
func (wxapi *WxApi) SendCustomMessage(ctx context.Context, openid string, reqBody map[string]any) error {
	_, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		reqBody["touser"] = openid
		req.SetBody(reqBody)
		return req.Post("/cgi-bin/message/custom/send")
	})

	if err != nil {
		return err
	}

	return nil
}

// 发送客服输入中
func (wxapi *WxApi) SendCustomTyping(ctx context.Context, openid string) error {
	_, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		reqBody := map[string]any{
			"touser":  openid,
			"command": "Typing",
		}
		req.SetBody(reqBody)
		return req.Post("/cgi-bin/message/custom/typing")
	})

	if err != nil {
		return err
	}

	return nil
}
