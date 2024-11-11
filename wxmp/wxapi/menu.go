package wxapi

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/go-resty/resty/v2"
)

type MenuResponse struct {
	IsMenuOpen int `json:"is_menu_open"`
	SelfMenu   struct {
		Button []*MenuButtonItem `json:"button"`
	} `json:"selfmenu_info"`
}

type MenuButtonItem struct {
	Type      string `json:"type"`
	Name      string `json:"name"`
	Value     string `json:"value,omitempty"`
	Key       string `json:"key,omitempty"`
	Url       string `json:"url,omitempty"`
	SubButton *struct {
		List []*MenuButtonItem `json:"list,omitempty"`
	} `json:"sub_button,omitempty"`
	NewsInfo *struct {
		List []struct {
			Title      string `json:"title"`
			Author     string `json:"author"`
			Digest     string `json:"digest"`
			ShowCover  int    `json:"show_cover"`
			CoverUrl   string `json:"cover_url"`
			ContentUrl string `json:"content_url"`
			SourceUrl  string `json:"source_url"`
		} `json:"list,omitempty"`
	} `json:"news_info,omitempty"`
}

type MenuButtonItemApiFormat struct {
	Type      string                     `json:"type"` // view click scancode_push scancode_waitmsg pic_sysphoto pic_photo_or_album pic_weixin location_select media_id article_id article_view_limited
	Name      string                     `json:"name"`
	Value     string                     `json:"value,omitempty"` // 调用查询菜单接口，才会有这个字段。创建时不会有
	Key       string                     `json:"key,omitempty"`
	Url       string                     `json:"url,omitempty"`
	MediaId   string                     `json:"media_id,omitempty"`
	ArticleId string                     `json:"article_id,omitempty"`
	ReplyData string                     `json:"reply_data,omitempty"`
	SubButton []*MenuButtonItemApiFormat `json:"sub_button,omitempty"`
	NewsInfo  *struct {
		List []struct {
			Title      string `json:"title"`
			Author     string `json:"author"`
			Digest     string `json:"digest"`
			ShowCover  int    `json:"show_cover"`
			CoverUrl   string `json:"cover_url"`
			ContentUrl string `json:"content_url"`
			SourceUrl  string `json:"source_url"`
		} `json:"list,omitempty"`
	} `json:"news_info,omitempty"`

	AppId    string `json:"appid,omitempty"`    // type为miniprogram时需要
	PagePath string `json:"pagepath,omitempty"` // type为miniprogram时需要
}

type MenuMatchRule struct {
	TagId              any    `json:"tag_id,omitempty"`
	ClientPlatformType string `json:"client_platform_type,omitempty"`
}

type AllMenuResponse struct {
	Menu struct {
		Button []*MenuButtonItemApiFormat `json:"button"`
		MenuId int                        `json:"menuid"`
	} `json:"menu"`
	Conditionalmenu []struct {
		Button    []*MenuButtonItemApiFormat `json:"button"`
		Matchrule *MenuMatchRule             `json:"matchrule"`
		MenuId    int                        `json:"menuid"`
	} `json:"conditionalmenu"`
}

func (wxapi *WxApi) GetMenu() (*MenuResponse, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		return req.Get("/cgi-bin/get_current_selfmenu_info")
	})
	if err != nil {
		return nil, err
	}

	// var err error
	// body := []byte(`
	// 		{
	//     "is_menu_open": 1,
	//     "selfmenu_info": {
	//         "button": [
	//             {
	//                 "name": "button",
	//                 "sub_button": {
	//                     "list": [
	//                         {
	//                             "type": "view",
	//                             "name": "view_url",
	//                             "url": "http://www.qq.com"
	//                         },
	//                         {
	//                             "type": "news",
	//                             "name": "news",
	//                             "key": "",
	//                             "value": "KQb_w_Tiz-nSdVLoTV35Psmty8hGBulGhEdbb9SKs-o",
	//                             "news_info": {
	//                                 "list": [
	//                                     {
	//                                         "title": "MULTI_NEWS",
	//                                         "author": "JIMZHENG",
	//                                         "digest": "text",
	//                                         "show_cover": 0,
	//                                         "cover_url": "http://mmbiz.qpic.cn/mmbiz/GE7et87vE9vicuCibqXsX9GPPLuEtBfXfK0HKuBIa1A1cypS0uY1wickv70iaY1gf3I1DTszuJoS3lAVLvhTcm9sDA/0",
	//                                         "content_url": "http://mp.weixin.qq.com/s?__biz=MjM5ODUwNTM3Ng==&mid=204013432&idx=1&sn=80ce6d9abcb832237bf86c87e50fda15#rd",
	//                                         "source_url": ""
	//                                     },
	//                                     {
	//                                         "title": "MULTI_NEWS1",
	//                                         "author": "JIMZHENG",
	//                                         "digest": "MULTI_NEWS1",
	//                                         "show_cover": 1,
	//                                         "cover_url": "http://mmbiz.qpic.cn/mmbiz/GE7et87vE9vicuCibqXsX9GPPLuEtBfXfKnmnpXYgWmQD5gXUrEApIYBCgvh2yHsu3ic3anDUGtUCHwjiaEC5bicd7A/0",
	//                                         "content_url": "http://mp.weixin.qq.com/s?__biz=MjM5ODUwNTM3Ng==&mid=204013432&idx=2&sn=8226843afb14ecdecb08d9ce46bc1d37#rd",
	//                                         "source_url": ""
	//                                     }
	//                                 ]
	//                             }
	//                         },
	//                         {
	//                             "type": "video",
	//                             "name": "video",
	//                             "value": "http://61.182.130.30/vweixinp.tc.qq.com/1007_114bcede9a2244eeb5ab7f76d951df5f.f10.mp4?vkey=77A42D0C2015FBB0A3653D29C571B5F4BBF1D243FBEF17F09C24FF1F2F22E30881BD350E360BC53F&sha=0&save=1"
	//                         },
	//                         {
	//                             "type": "voice",
	// 							"key": "key_2",
	//                             "name": "voice",
	//                             "value": "2nG3qG0FnNnv11IiVJlgn8yWm01RVy_no9MG3V0sRUhIbK1Wcywaig2RN8i4SGxT"
	//                         }
	//                     ]
	//                 }
	//             },
	//             {
	//                 "type": "text",
	//                 "name": "text",
	//                 "value": "This is text!"
	//             },
	//             {
	//                 "type": "img",
	//                 "name": "photo",
	//                 "value": "2nG3qG0FnNnv11IiVJlgn0pbbHc-DCGIZ6O_4_xxbqhKo7PW6tkOwtNAxPKfAeLS"
	//             }
	//         ]
	//     }
	// }
	// 	`)

	var ret MenuResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}

// 创建普通菜单
func (wxapi *WxApi) CreateMenu(buttons []*MenuButtonItemApiFormat) (string, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"button": buttons,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/menu/create")
	})
	if err != nil {
		return "", err
	}

	log.Println("CreateMenu response:", string(body))

	return string(body), nil
}

// 创建个性化菜单
func (wxapi *WxApi) CreateMenuConditional(buttons []*MenuButtonItemApiFormat, matchrule *MenuMatchRule) (string, error) {
	log.Println("CreateMenuConditional matchrule:", matchrule)
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"button":    buttons,
			"matchrule": matchrule,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/menu/addconditional")
	})
	if err != nil {
		return "", err
	}

	log.Println("CreateMenuConditional response:", string(body))

	ret := struct {
		MenuId int `json:"menuid"`
	}{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(ret.MenuId), nil
}

// 调用此接口会删除默认菜单及全部个性化菜单
func (wxapi *WxApi) DeleteMenu(ctx context.Context) error {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		return req.Get("/cgi-bin/menu/delete")
	})
	if err != nil {
		return err
	}

	log.Println("DeleteMenu response:", string(body))

	return nil
}

// 删除个性化菜单
func (wxapi *WxApi) DeleteMenuConditional(menuId string) error {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"menuid": menuId,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/menu/delconditional")
	})
	if err != nil {
		return err
	}

	log.Println("DeleteMenuConditional response:", string(body))

	return nil
}

// 测试个性化菜单匹配结果
func (wxapi *WxApi) TryMatchMenu(userOpenId string) ([]*MenuButtonItemApiFormat, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"user_id": userOpenId,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/menu/trymatch")
	})
	if err != nil {
		return nil, err
	}

	log.Println("TryMatchMenu response:", string(body))

	ret := struct {
		Button []*MenuButtonItemApiFormat `json:"button"`
	}{}
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return ret.Button, nil
}

// 获取所有菜单
func (wxapi *WxApi) GetAllMenu() (*AllMenuResponse, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		return req.Get("/cgi-bin/menu/get")
	})
	if err != nil {
		return nil, err
	}

	// log.Println("GetAllMenu response:", string(body))

	var ret AllMenuResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}
	return &ret, nil
}
