package controllers

import (
	"github.com/anchel/wechat-official-account-admin/modules/weixin"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-gonic/gin"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &WeixinUserController{
			BaseController: &BaseController{},
		}
		r.GET("/wxuser/tag/list", ctl.GetTagList)
		r.POST("/wxuser/tag/create", ctl.CreateTag)
		r.POST("/wxuser/tag/update", ctl.UpdateTag)
		r.POST("/wxuser/tag/delete", ctl.DeleteTag)
		r.GET("/wxuser/tag/users", ctl.GetTagUsers)

		r.POST("/wxuser/tag/batchtagging", ctl.BatchTagging)
		r.POST("/wxuser/tag/batchuntagging", ctl.BatchUntagging)
		r.GET("/wxuser/tag/getusertags", ctl.GetUserTags)

		r.GET("/wxuser/list", ctl.GetUserList)
		r.POST("/wxuser/set-remark", ctl.UpdateUserRemark)
	})
}

type WeixinUserController struct {
	*BaseController
}

// 获取标签列表
func (ctl *WeixinUserController) GetTagList(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	list, err := wxApiClient.GetTagList(ctx)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, gin.H{"list": list})
}

// 创建用户标签
func (ctl *WeixinUserController) CreateTag(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	id, err := wxApiClient.CreateTag(ctx, form.Name)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, gin.H{"id": id})
}

// 更新标签
func (ctl *WeixinUserController) UpdateTag(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		ID   int    `json:"id"`
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = wxApiClient.UpdateTag(ctx, form.ID, form.Name)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, nil)
}

// 删除标签
func (ctl *WeixinUserController) DeleteTag(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		ID int `json:"id"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = wxApiClient.DeleteTag(ctx, form.ID)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, nil)
}

// 获取标签下粉丝列表
func (ctl *WeixinUserController) GetTagUsers(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		TagID      int    `json:"tagid" form:"tagid" binding:"required"`
		NextOpenid string `json:"next_openid" form:"next_openid"`
	}
	if err := c.ShouldBindQuery(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	count, list, next_openid, err := wxApiClient.GetTagUsers(ctx, form.TagID, form.NextOpenid)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, gin.H{"total": 0, "count": count, "list": list, "next_openid": next_openid})
}

// 批量为用户打标签
func (ctl *WeixinUserController) BatchTagging(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		TagID      int      `json:"tagid" binding:"required"`
		OpenidList []string `json:"openid_list" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = wxApiClient.BatchTagging(ctx, form.TagID, form.OpenidList)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, nil)
}

// 批量为用户取消标签
func (ctl *WeixinUserController) BatchUntagging(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		TagID      int      `json:"tagid" binding:"required"`
		OpenidList []string `json:"openid_list" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = wxApiClient.BatchUntagging(ctx, form.TagID, form.OpenidList)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, nil)
}

// 获取用户身上的标签列表
func (ctl *WeixinUserController) GetUserTags(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		Openid string `json:"openid" form:"openid" binding:"required"`
	}
	if err := c.ShouldBindQuery(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	list, err := wxApiClient.GetUserTags(ctx, form.Openid)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, gin.H{"list": list})
}

// 获取用户列表
func (ctl *WeixinUserController) GetUserList(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		NextOpenid string `json:"next_openid" form:"next_openid"`
	}
	if err := c.ShouldBindQuery(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	total, list, next_openid, err := wxApiClient.GetUserList(ctx, form.NextOpenid)
	if ctl.checkError(c, err) != nil {
		return
	}

	// users := make([]*wxapi.UserInfo, len(list))

	// 批量获取用户信息
	userInfoList, err := wxApiClient.BatchGetUserInfo(ctx, list)
	if ctl.checkError(c, err) != nil {
		return
	}

	// var wg sync.WaitGroup

	// for i := range list {
	// 	wg.Add(1)
	// 	go func(i int) {
	// 		defer wg.Done()
	// 		var tags []int
	// 		tags, err = wxApiClient.GetUserTags(ctx, list[i])
	// 		if err != nil {
	// 			log.Println("GetUserTags", err.Error())
	// 			return
	// 		}
	// 		users[i] = &wxapi.UserInfo{
	// 			Openid:    list[i],
	// 			TagIdList: tags,
	// 		}
	// 	}(i)
	// }
	// wg.Wait()

	// if ctl.checkError(c, err) != nil {
	// 	return
	// }

	ctl.returnOk(c, gin.H{"total": total, "list": userInfoList, "next_openid": next_openid})
}

// 设置用户备注名
func (ctl *WeixinUserController) UpdateUserRemark(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var form struct {
		Openid string `json:"openid" form:"openid" binding:"required"`
		Remark string `json:"remark" form:"remark" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = wxApiClient.UpdateUserRemark(ctx, form.Openid, form.Remark)
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, nil)
}
