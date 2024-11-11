package wxapi

import (
	"context"
	"encoding/json"
	"log"

	"github.com/go-resty/resty/v2"
)

type UserMgrTag struct {
	Id    int    `json:"id"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// 获取标签列表
func (wxapi *WxApi) GetTagList(ctx context.Context) ([]UserMgrTag, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		return req.Get("/cgi-bin/tags/get")
	})
	if err != nil {
		return nil, err
	}

	retobj := struct {
		Tags []UserMgrTag `json:"tags"`
	}{}
	err = json.Unmarshal(body, &retobj)
	if err != nil {
		log.Println("GetTagList json.Unmarshal", err.Error())
		return nil, err
	}

	return retobj.Tags, nil
}

// 创建标签
func (wxapi *WxApi) CreateTag(ctx context.Context, name string) (int, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"tag": map[string]string{
				"name": name,
			},
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/tags/create")
	})
	if err != nil {
		return 0, err
	}

	retobj := struct {
		Tag struct {
			Id int `json:"id"`
		} `json:"tag"`
	}{}
	err = json.Unmarshal(body, &retobj)
	if err != nil {
		log.Println("CreateTag json.Unmarshal", err.Error())
		return 0, err
	}

	return retobj.Tag.Id, nil
}

// 修改标签
func (wxapi *WxApi) UpdateTag(ctx context.Context, id int, name string) error {
	_, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"tag": map[string]interface{}{
				"id":   id,
				"name": name,
			},
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/tags/update")
	})
	if err != nil {
		return err
	}

	return nil
}

// 删除标签
func (wxapi *WxApi) DeleteTag(ctx context.Context, id int) error {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"tag": map[string]interface{}{
				"id": id,
			},
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/tags/delete")
	})
	if err != nil {
		log.Println("DeleteTag", string(body))
		return err
	}

	log.Println("DeleteTag", string(body))

	return nil
}

// 获取标签下粉丝列表
func (wxapi *WxApi) GetTagUsers(ctx context.Context, tagid int, next_openid string) (int, []string, string, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"tagid":       tagid,
			"next_openid": next_openid,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/user/tag/get")
	})
	if err != nil {
		return 0, nil, "", err
	}

	log.Println("GetTagUsers", string(body))

	retobj := struct {
		Count int `json:"count"`
		Data  struct {
			Openid []string `json:"openid"`
		} `json:"data"`
		NextOpenid string `json:"next_openid"`
	}{}
	err = json.Unmarshal(body, &retobj)
	if err != nil {
		log.Println("GetTagUsers json.Unmarshal", err.Error())
		return 0, nil, "", err
	}

	return retobj.Count, retobj.Data.Openid, retobj.NextOpenid, nil
}

// 批量为用户打标签
func (wxapi *WxApi) BatchTagging(ctx context.Context, tagid int, openid_list []string) error {
	_, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"tagid":       tagid,
			"openid_list": openid_list,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/tags/members/batchtagging")
	})
	if err != nil {
		return err
	}

	return nil
}

// 批量为用户取消标签
func (wxapi *WxApi) BatchUntagging(ctx context.Context, tagid int, openid_list []string) error {
	_, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"tagid":       tagid,
			"openid_list": openid_list,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/tags/members/batchuntagging")
	})
	if err != nil {
		return err
	}

	return nil
}

// 获取用户身上的标签列表
func (wxapi *WxApi) GetUserTags(ctx context.Context, openid string) ([]int, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"openid": openid,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/tags/getidlist")
	})
	if err != nil {
		return nil, err
	}

	retobj := struct {
		TagidList []int `json:"tagid_list"`
	}{}
	err = json.Unmarshal(body, &retobj)
	if err != nil {
		log.Println("GetUserTags json.Unmarshal", err.Error())
		return nil, err
	}

	return retobj.TagidList, nil
}

// 设置用户备注名
func (wxapi *WxApi) UpdateUserRemark(ctx context.Context, openid, remark string) error {
	_, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"openid": openid,
			"remark": remark,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/user/info/updateremark")
	})
	if err != nil {
		return err
	}

	return nil
}

type UserInfo struct {
	Subscribe      int    `json:"subscribe"`
	Openid         string `json:"openid"`
	Unionid        string `json:"unionid"`
	Language       string `json:"language"`
	SubscribeTime  int    `json:"subscribe_time"`
	Remark         string `json:"remark"`
	Groupid        int    `json:"groupid"`
	TagIdList      []int  `json:"tagid_list"`
	SubscribeScene string `json:"subscribe_scene"`
	QrScene        int    `json:"qr_scene"`
	QrSceneStr     string `json:"qr_scene_str"`
}

// 获取用户基本信息
func (wxapi *WxApi) GetUserInfo(ctx context.Context, openid string) (*UserInfo, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		req.SetQueryParam("openid", openid)
		req.SetQueryParam("lang", "zh_CN")
		return req.Get("/cgi-bin/user/info")
	})
	if err != nil {
		return nil, err
	}

	retobj := &UserInfo{}
	err = json.Unmarshal(body, retobj)
	if err != nil {
		log.Println("GetUserInfo json.Unmarshal", err.Error())
		return nil, err
	}

	return retobj, nil
}

// 批量获取用户基本信息
func (wxapi *WxApi) BatchGetUserInfo(ctx context.Context, openid_list []string) ([]UserInfo, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"user_list": []map[string]string{},
		}
		for _, openid := range openid_list {
			body["user_list"] = append(body["user_list"].([]map[string]string), map[string]string{"openid": openid, "lang": "zh_CN"})
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/user/info/batchget")
	})
	if err != nil {
		return nil, err
	}

	retobj := struct {
		UserInfoList []UserInfo `json:"user_info_list"`
	}{}
	err = json.Unmarshal(body, &retobj)
	if err != nil {
		log.Println("BatchGetUserInfo json.Unmarshal", err.Error())
		return nil, err
	}

	return retobj.UserInfoList, nil
}

// 获取用户列表
func (wxapi *WxApi) GetUserList(ctx context.Context, next_openid string) (int, []string, string, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		req.SetQueryParam("next_openid", next_openid)
		return req.Get("/cgi-bin/user/get")
	})
	if err != nil {
		return 0, nil, "", err
	}

	retobj := struct {
		Total int `json:"total"`
		Count int `json:"count"`
		Data  struct {
			Openid []string `json:"openid"`
		} `json:"data"`
		NextOpenid string `json:"next_openid"`
	}{}
	err = json.Unmarshal(body, &retobj)
	if err != nil {
		log.Println("GetUserList json.Unmarshal", err.Error())
		return 0, nil, "", err
	}

	return retobj.Total, retobj.Data.Openid, retobj.NextOpenid, nil
}
