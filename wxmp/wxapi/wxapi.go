package wxapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	accesstokenstore "github.com/anchel/wechat-official-account-admin/wxmp/access-token-store"
	mpoptions "github.com/anchel/wechat-official-account-admin/wxmp/mp-options"
	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
)

type WxApi struct {
	mpoptions   *mpoptions.MpOptions
	mpstore     accesstokenstore.AccessTokenStore
	restyClient *resty.Client
}

func NewWxApi(mpoptions *mpoptions.MpOptions, rdb *redis.Client) *WxApi {
	mpstore := accesstokenstore.NewRedisStore(mpoptions, rdb)
	restyClient := resty.New()
	wxProxy := os.Getenv("WA_PROXY")
	if wxProxy != "" {
		log.Println("WA_PROXY:", wxProxy)
		restyClient.SetProxy(wxProxy)
	}
	restyClient.SetBaseURL("https://api.weixin.qq.com")
	return &WxApi{mpoptions: mpoptions, mpstore: mpstore, restyClient: restyClient}
}

func (wxapi *WxApi) GetMpOptions() *mpoptions.MpOptions {
	return wxapi.mpoptions
}

func (wxapi *WxApi) GetAppID() string {
	return wxapi.mpoptions.AppId
}

func (wxapi *WxApi) NewRequest() (*resty.Request, error) {
	req := wxapi.restyClient.R()
	token, err := wxapi.mpstore.Get()
	if err != nil {
		log.Println("NewRequest wxapi.mpstore.Get()", err.Error())
		return nil, err
	}
	req.SetQueryParam("access_token", token)
	return req, nil
}

func (wxapi *WxApi) PrepareRequest(req *resty.Request) error {
	token, err := wxapi.mpstore.Get()
	if err != nil {
		log.Println("PrepareRequest wxapi.mpstore.Get()", err.Error())
		return err
	}
	req.SetQueryParam("access_token", token)
	return nil
}

func (wxapi *WxApi) PrepareResponse(res *resty.Response, checkErrcode bool) ([]byte, *resty.Response, error) {
	if res == nil {
		return nil, res, errors.New("response is nil")
	}

	if res.StatusCode() != 200 {
		return nil, res, errors.New("http status code not 200")
	}
	body := res.Body()

	contentType := res.Header().Get("Content-Type")
	log.Println("PrepareResponse contentType:", contentType, len(body))
	// 如果返回的是文件，则直接返回，不检查errcode
	if res.Header().Get("Content-Disposition") != "" {
		return body, res, nil
	}

	if checkErrcode && body != nil {
		retobj := struct {
			Errcode int64  `json:"errcode"`
			Errmsg  string `json:"errmsg"`
		}{}
		err := json.Unmarshal(body, &retobj)
		if err != nil {
			return nil, res, err
		}
		if retobj.Errcode != 0 {
			log.Println("PrepareResponse errcode:", retobj.Errcode, retobj.Errmsg)
			if retobj.Errcode == 40001 || retobj.Errcode == 40014 {
				retobj.Errmsg = fmt.Sprint("access_token 无效", "-", retobj.Errmsg)
				_, err := wxapi.mpstore.Refresh(true)
				if err != nil {
					log.Println("PrepareResponse wxapi.mpstore.Refresh()", err.Error())
					return nil, res, err
				}
			}

			if retobj.Errcode == 48001 {
				retobj.Errmsg = fmt.Sprint("api 功能未授权，请确认公众号已获得该接口权限", "-", retobj.Errmsg)
			}
			return nil, res, errors.New(fmt.Sprint(retobj.Errcode, "-", retobj.Errmsg))
		}
	}
	return body, res, nil
}

type ModifyRequestFunc func(req *resty.Request) (*resty.Response, error)

func (wxapi *WxApi) CommonRequest(checkErrcode bool, modifier ModifyRequestFunc) ([]byte, *resty.Response, error) {
	req, err := wxapi.NewRequest()
	if err != nil {
		return nil, nil, err
	}
	var res *resty.Response
	if modifier != nil {
		res, err = modifier(req)
		if err != nil {
			return nil, res, err
		}
	}
	return wxapi.PrepareResponse(res, checkErrcode)
}
