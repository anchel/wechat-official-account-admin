package controllers

import (
	"github.com/anchel/wechat-official-account-admin/modules/weixin"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/anchel/wechat-official-account-admin/wxmp/wxapi"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &QrcodeController{
			BaseController: &BaseController{},
		}
		r.GET("/qrcode/list", ctl.List)
		r.POST("/qrcode/add", ctl.Create)
	})
}

type QrcodeController struct {
	*BaseController
}

// 获取二维码列表
func (ctl *QrcodeController) List(c *gin.Context) {
	var form struct {
		QrcodeType string `json:"qrcode_type" form:"qrcode_type" binding:"required"` // temp-临时，limit-永久
		Offset     *int64 `json:"offset" form:"offset" binding:"required"`
		Count      *int64 `json:"count" form:"count" binding:"required"`
		Keyword    string `json:"keyword" form:"keyword"`
	}
	if err := c.ShouldBindQuery(&form); err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	if !lo.Contains([]string{"temp", "limit"}, form.QrcodeType) {
		ctl.returnFail(c, 1, "qrcode_type must be temp or limit")
		return
	}

	// 从数据库中查询数据，查询title或scene_str包含search_keyword的数据，按创建时间倒序排列，分页返回
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	findOptions := options.Find()
	findOptions.SetSkip(*form.Offset)
	findOptions.SetLimit(*form.Count)
	findOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	filter := bson.D{
		{Key: "appid", Value: appid},
		{Key: "qrcode_type", Value: form.QrcodeType},
		{Key: "$or", Value: bson.A{bson.D{{Key: "title", Value: bson.D{{Key: "$regex", Value: form.Keyword}}}}, bson.D{{Key: "scene_str", Value: bson.D{{Key: "$regex", Value: form.Keyword}}}}}},
	}

	total, err := mongodb.ModelWxQrcode.Count(ctx, filter)
	if ctl.checkError(c, err) != nil {
		return
	}

	docs, err := mongodb.ModelWxQrcode.FindMany(ctx, filter, findOptions)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, gin.H{"total": total, "list": docs})
}

// 创建永久二维码或临时二维码
func (ctl *QrcodeController) Create(c *gin.Context) {
	var form struct {
		QrcodeType    string `json:"qrcode_type" form:"qrcode_type" binding:"required"` // temp-临时，limit-永久
		Title         string `json:"title" form:"title" binding:"required"`
		SceneStr      string `json:"scene_str" form:"scene_str"`
		SceneId       int    `json:"scene_id" form:"scene_id"`
		ExpireSeconds int    `json:"expire_seconds" form:"expire_seconds"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	if !lo.Contains([]string{"temp", "limit"}, form.QrcodeType) {
		ctl.returnFail(c, 1, "qrcode_type must be temp or limit")
		return
	}

	if form.SceneStr == "" && form.SceneId == 0 {
		ctl.returnFail(c, 1, "scene_str or scene_id is required")
		return
	}

	var ret *wxapi.CreateQrCodeResp
	var err error

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	if form.QrcodeType == "temp" {
		ret, err = wxApiClient.CreateTempQrCode(ctx, form.SceneStr, form.SceneId, form.ExpireSeconds)
	} else if form.QrcodeType == "limit" {
		ret, err = wxApiClient.CreateQrCode(ctx, form.SceneStr, form.SceneId)
	}

	if ctl.checkError(c, err) != nil {
		return
	}

	doc := &mongodb.EntityWxQrcode{
		AppID:         appid,
		QrcodeType:    form.QrcodeType,
		SceneStr:      form.SceneStr,
		SceneId:       form.SceneId,
		Title:         form.Title,
		Ticket:        ret.Ticket,
		ExpireSeconds: ret.ExpireSeconds,
		Url:           ret.Url,
	}
	id, err := mongodb.ModelWxQrcode.InsertOne(ctx, doc)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, gin.H{"id": id, "ticket": ret.Ticket, "url": ret.Url, "expire_seconds": ret.ExpireSeconds})
}
