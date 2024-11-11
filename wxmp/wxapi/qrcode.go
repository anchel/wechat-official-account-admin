package wxapi

import (
	"context"
	"encoding/json"

	"github.com/go-resty/resty/v2"
)

type CreateQrCodeResp struct {
	Ticket        string `json:"ticket"`
	ExpireSeconds int    `json:"expire_seconds"`
	Url           string `json:"url"`
}

// 创建永久二维码
func (wxapi *WxApi) CreateQrCode(ctx context.Context, sceneStr string, sceneId int) (*CreateQrCodeResp, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {

		if sceneStr != "" {
			body := map[string]interface{}{
				"action_name": "QR_LIMIT_STR_SCENE",
				"action_info": map[string]interface{}{
					"scene": map[string]string{
						"scene_str": sceneStr,
					},
				},
			}

			req.SetBody(body)
		} else {
			body := map[string]interface{}{
				"action_name": "QR_LIMIT_SCENE",
				"action_info": map[string]interface{}{
					"scene": map[string]int{
						"scene_id": sceneId,
					},
				},
			}

			req.SetBody(body)
		}
		return req.Post("/cgi-bin/qrcode/create")
	})

	if err != nil {
		return nil, err
	}

	retobj := &CreateQrCodeResp{}
	err = json.Unmarshal(body, retobj)
	if err != nil {
		return nil, err
	}

	return retobj, nil
}

// 创建临时二维码
func (wxapi *WxApi) CreateTempQrCode(ctx context.Context, sceneStr string, sceneId int, expireSeconds int) (*CreateQrCodeResp, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		if expireSeconds > 2592000 {
			expireSeconds = 2592000
		}
		if sceneStr != "" {
			body := map[string]interface{}{
				"expire_seconds": expireSeconds,
				"action_name":    "QR_STR_SCENE",
				"action_info": map[string]interface{}{
					"scene": map[string]string{
						"scene_str": sceneStr,
					},
				},
			}

			req.SetBody(body)
		} else {
			body := map[string]interface{}{
				"expire_seconds": expireSeconds,
				"action_name":    "QR_SCENE",
				"action_info": map[string]interface{}{
					"scene": map[string]int{
						"scene_id": sceneId,
					},
				},
			}
			req.SetBody(body)
		}
		return req.Post("/cgi-bin/qrcode/create")
	})

	if err != nil {
		return nil, err
	}

	retobj := &CreateQrCodeResp{}
	err = json.Unmarshal(body, retobj)
	if err != nil {
		return nil, err
	}

	return retobj, nil
}
