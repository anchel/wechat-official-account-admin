package controllers

import (
	"context"
	"fmt"
	"strings"

	"log"

	"github.com/anchel/wechat-official-account-admin/lib/types"
	"github.com/anchel/wechat-official-account-admin/lib/utils"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	appidservice "github.com/anchel/wechat-official-account-admin/services/appid-service"
	commonservice "github.com/anchel/wechat-official-account-admin/services/common-service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &AppIDController{
			BaseController: &BaseController{},
		}
		// 下面是不需要选择了某个公众号就可以调用的
		r.GET("/system/appid/list", ctl.List)
		r.GET("/system/appid/get_config_info", ctl.GetConfigInfo)
		r.POST("/system/appid/select", ctl.Select)
		r.POST("/system/appid/save", ctl.Save)
		r.POST("/system/appid/delete", ctl.Delete)

		// 下面是需要已经选择了某个公众号的
		r.GET("/appid/session_info", ctl.SessionInfo)
		r.GET("/appid/get_enabled", ctl.GetEnabled)
		r.POST("/appid/set_enabled", ctl.SetEnabled)
	})
}

type AppIDController struct {
	*BaseController
}

// 拉取公众号列表
func (ctl *AppIDController) List(c *gin.Context) {
	ctx := context.Background()

	// 如果是管理员，返回appsecret、token、encoding_aes_key等敏感信息
	isAdmin, err := ctl.checkIsAdmin(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	list := make([]*mongodb.EntityWxAppid, 0)
	appidList, err := appidservice.GetAppIDList(ctx)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	// log.Println("c.Request.URL", c.Request.URL.Scheme, c.Request.Header.Get("X-Forwarded-Proto"))
	protocol := c.Request.URL.Scheme
	if protocol == "" {
		protocol = c.Request.Header.Get("X-Forwarded-Proto")
	}
	if protocol == "" {
		protocol = "http"
	}
	protocol, _ = strings.CutSuffix(protocol, "://")

	for _, app := range appidList {
		if !isAdmin {
			app.AppSecret = ""
			app.Token = ""
			app.EncodingAESKey = ""
		}
		if app.Thumbnail != "" && !strings.HasPrefix(app.Thumbnail, "http") {
			app.Thumbnail = fmt.Sprint(protocol, "://", app.Thumbnail)
		}
		list = append(list, app)
	}

	ctl.returnOk(c, gin.H{"list": list})
}

// 选择当前管理的公众号
func (ctl *AppIDController) Select(c *gin.Context) {
	ctx := context.Background()

	var form struct {
		AppID string `json:"appid" form:"appid" binding:"required"`
	}
	if c.ShouldBindJSON(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	appidList, err := appidservice.GetAppIDList(ctx)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	// 判断提交的appid是否在列表中
	app, extsts := lo.Find(appidList, func(app *mongodb.EntityWxAppid) bool {
		return app.AppID == form.AppID
	})
	if !extsts {
		ctl.returnFail(c, 400, "appid不存在")
		return
	}

	log.Println("select appid", app.AppID)
	sessionAppidInfo := types.SessionAppidInfo{
		AppType:        app.AppType,
		Name:           app.Name,
		AppID:          app.AppID,
		AppSecret:      app.AppSecret,
		Token:          app.Token,
		EncodingAESKey: app.EncodingAESKey,
	}

	session := sessions.Default(c)
	session.Set("appid", sessionAppidInfo)
	if err := session.Save(); err != nil {
		log.Println("session save fail", err)
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"appid": app.AppID})
}

// 获取当前登录态的appid信息
func (ctl *AppIDController) SessionInfo(c *gin.Context) {
	session := sessions.Default(c)
	app, ok := session.Get("appid").(types.SessionAppidInfo)
	if ok {
		ctl.returnOk(c, gin.H{"appidInfo": app})
	} else {
		ctl.returnFail(c, 1, "no appid")
	}
}

// 保存公众号
func (ctl *AppIDController) Save(c *gin.Context) {
	var form struct {
		ID             string `json:"id" form:"id"`                                  // 公众号id
		AppType        string `json:"app_type" form:"app_type" binding:"required"`   // 公众号类型
		Name           string `json:"name" form:"name" binding:"required"`           // 公众号名称
		AppID          string `json:"appid" form:"appid" binding:"required"`         // 公众号appid
		AppSecret      string `json:"appsecret" form:"appsecret" binding:"required"` // 公众号appsecret
		Token          string `json:"token" form:"token"`                            // 公众号token
		EncodingAESKey string `json:"encoding_aes_key" form:"encoding_aes_key"`      // 公众号encoding_aes_key
		Thumbnail      string `json:"thumbnail" form:"thumbnail"`                    // 公众号缩略图
	}
	// token encoding_aes_key是可选的，是因为有些只需要调用接口，不需要接收消息

	if c.ShouldBindJSON(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}
	ctx := context.Background()

	// 检查当前登录用户是否有权限
	isAdmin, err := ctl.checkIsAdmin(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	if !isAdmin {
		ctl.returnFail(c, 1, "no permission")
		return
	}

	// 如果有id，就是更新
	if form.ID != "" {
		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "app_type", Value: form.AppType},
				{Key: "name", Value: form.Name},
				{Key: "appsecret", Value: form.AppSecret},
				{Key: "token", Value: form.Token},
				{Key: "encoding_aes_key", Value: form.EncodingAESKey},
				{Key: "thumbnail", Value: form.Thumbnail},
			}},
		}
		ret, err := mongodb.ModelWxAppid.UpdateByID(ctx, form.ID, update)
		if err != nil {
			ctl.returnFail(c, 500, err.Error())
			return
		}
		ctl.returnOk(c, gin.H{"id": form.ID, "result": ret})
		return
	}

	// 如果没有id，就是新增
	doc := &mongodb.EntityWxAppid{
		AppType:        form.AppType,
		Name:           form.Name,
		AppID:          form.AppID,
		AppSecret:      form.AppSecret,
		Token:          form.Token,
		EncodingAESKey: form.EncodingAESKey,
		Thumbnail:      form.Thumbnail,
	}
	id, err := mongodb.ModelWxAppid.InsertOne(ctx, doc)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"id": id, "appid": form.AppID})
}

// 删除公众号
func (ctl *AppIDController) Delete(c *gin.Context) {
	var form struct {
		ID string `json:"id" form:"id" binding:"required"`
	}
	if c.ShouldBindJSON(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}
	ctx := context.Background()

	// 检查当前登录用户是否有权限
	isAdmin, err := ctl.checkIsAdmin(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	if !isAdmin {
		ctl.returnFail(c, 1, "no permission")
		return
	}

	ret, err := mongodb.ModelWxAppid.DeleteByID(ctx, form.ID)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"id": form.ID, "result": ret})
}

// 获取公众号功能的启用状态
func (ctl *AppIDController) GetEnabled(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ret, err := appidservice.GetAppEnabledData(ctx, appid)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	ctl.returnOk(c, ret)
}

// 设置公众号功能的启用状态
func (ctl *AppIDController) SetEnabled(c *gin.Context) {
	var form struct {
		ReplyType string `json:"reply_type" form:"reply_type" binding:"required"` // subscribe, keyword, message
		Enabled   *bool  `json:"enabled" form:"enabled" binding:"required"`
	}
	if c.ShouldBindJSON(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	err = appidservice.SetAppEnabledData(ctx, appid, form.ReplyType, *form.Enabled)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	ctl.returnOk(c, nil)
}

// 获取公众号的附加信息
func (ctl *AppIDController) GetConfigInfo(c *gin.Context) {
	appid := c.Query("appid")
	if appid == "" {
		ctl.returnFail(c, 400, "appid required")
		return
	}

	ip, err := commonservice.GetPublicIP()
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	pathname := fmt.Sprint("/wxmp/", appid, "/handler")
	ctl.returnOk(c, gin.H{"ip": ip, "configUrl": utils.MakePublicServeUrl(c, pathname)})
}
