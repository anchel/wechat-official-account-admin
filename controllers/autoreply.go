package controllers

import (
	"encoding/json"
	"log"
	"time"

	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	replyservice "github.com/anchel/wechat-official-account-admin/services/reply-service"
	weixinservice "github.com/anchel/wechat-official-account-admin/services/weixin-service"
	"github.com/gin-gonic/gin"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &AutoReplyController{
			BaseController: &BaseController{},
		}
		r.GET("/autoreply/get", ctl.Get)
		r.POST("/autoreply/save", ctl.Save)
		r.GET("/autoreply/delete", ctl.Delete)
	})
}

type AutoReplyController struct {
	*BaseController
}

type KeywordDefIntem struct {
	Keyword string `json:"keyword"`
	Exact   bool   `json:"exact"`
}

type AutoReplyGetRespItem struct {
	ID          string                       `json:"id"`
	ReplyType   string                       `json:"reply_type"`
	ReplyData   *weixinservice.AutoReplyData `json:"reply_data"`
	RuleTitle   string                       `json:"rule_title"`
	Keywords    []string                     `json:"keywords"`
	KeywordsDef []*KeywordDefIntem           `json:"keywords_def"`
	CreatedAt   time.Time                    `json:"created_at"`
}

// 查询关注回复、关键词回复、消息回复
func (ctl *AutoReplyController) Get(c *gin.Context) {
	var form struct {
		ReplyType string `json:"reply_type" form:"reply_type"` // subscribe, keyword, message
		Search    string `json:"search" form:"search"`
	}
	if c.ShouldBindQuery(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}
	flag := lo.Contains([]string{"subscribe", "keyword", "message"}, form.ReplyType)
	if !flag {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	_, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	findOptions := options.Find()
	filter := bson.D{{Key: "appid", Value: appid}, {Key: "reply_type", Value: form.ReplyType}}
	if form.Search != "" {
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{
			bson.D{{Key: "rule_title", Value: bson.D{{Key: "$regex", Value: form.Search}}}},
			bson.D{{Key: "keywords", Value: bson.D{{Key: "$regex", Value: form.Search}}}},
		}})
	}

	docs, err := mongodb.ModelWeixinAutoReply.FindMany(c, filter, findOptions)
	if err != nil {
		log.Println("mongodb.ModelWeixinAutoReply.FindMany error", err)
		ctl.returnFail(c, 500, "查询失败")
		return
	}

	list := []AutoReplyGetRespItem{}
	for _, doc := range docs {
		replyData, err := replyservice.ParseAutoReplyData(doc.ReplyData)
		if err != nil {
			ctl.returnFail(c, 500, "解析replydata失败:"+doc.ID.Hex())
			return
		}

		item := AutoReplyGetRespItem{
			ID:        doc.ID.Hex(),
			ReplyType: doc.ReplyType,
			ReplyData: replyData,
			RuleTitle: doc.RuleTitle,
			Keywords:  doc.Keywords,
			CreatedAt: doc.CreatedAt,
		}
		if doc.KeywordsDef != "" {
			item.KeywordsDef = []*KeywordDefIntem{}
			err = json.Unmarshal([]byte(doc.KeywordsDef), &item.KeywordsDef)
			if err != nil {
				ctl.returnFail(c, 500, "转换keywordsdef失败")
				return
			}
		}
		list = append(list, item)
	}

	ctl.returnOk(c, gin.H{"list": list})
}

type AutoReplySaveForm struct {
	ID        string                       `json:"id" form:"id"`
	ReplyType string                       `json:"reply_type" form:"reply_type" binding:"required"` // subscribe, keyword, message
	ReplyData *weixinservice.AutoReplyData `json:"reply_data" form:"reply_data" binding:"required"`

	RuleTitle   string             `json:"rule_title" form:"rule_title"`
	Keywords    []string           `json:"keywords" form:"keywords"`
	KeywordsDef []*KeywordDefIntem `json:"keywords_def" form:"keywords_def"`
}

// 保存关注回复、关键词回复、消息回复
// 这个和菜单点击有些区别。菜单点击有草稿。而这三种类型没有草稿
func (ctl *AutoReplyController) Save(c *gin.Context) {
	var form AutoReplySaveForm
	if c.ShouldBindJSON(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	if !lo.Contains([]string{"subscribe", "keyword", "message"}, form.ReplyType) {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	_, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	replyDataStr, err := json.Marshal(form.ReplyData)
	if err != nil {
		ctl.returnFail(c, 500, "转换replydata失败")
		return
	}
	keywordsDefStr, err := json.Marshal(form.KeywordsDef)
	if err != nil {
		ctl.returnFail(c, 500, "转换keywordsdef失败")
		return
	}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "reply_data", Value: string(replyDataStr)}}}}
	if form.ReplyType == "keyword" {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "rule_title", Value: form.RuleTitle}}})
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "keywords", Value: form.Keywords}}})
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "keywords_def", Value: string(keywordsDefStr)}}})
	}

	// 如果有ID，就是更新
	if form.ID != "" {
		_, err := mongodb.ModelWeixinAutoReply.UpdateByID(c, form.ID, update)
		if err != nil {
			ctl.returnFail(c, 500, err.Error())
			return
		}
		ctl.returnOk(c, gin.H{"id": form.ID})
		return
	}

	// 如果是关注回复或者消息回复，只能有一条
	if form.ReplyType == "message" || form.ReplyType == "subscribe" {
		filter := bson.D{{Key: "appid", Value: appid}, {Key: "reply_type", Value: form.ReplyType}}
		doc, err := mongodb.ModelWeixinAutoReply.FindOneAndUpdate(c, filter, update, true)
		if err != nil {
			ctl.returnFail(c, 500, err.Error())
			return
		}
		ctl.returnOk(c, gin.H{"id": doc.ID.Hex()})
		return
	}

	// 关键词回复可以有多条，所以不带ID的，都是新增
	doc := mongodb.EntityWeixinAutoReply{
		AppID:     appid,
		ReplyType: form.ReplyType,
		ReplyData: string(replyDataStr),

		RuleTitle:   form.RuleTitle,
		Keywords:    form.Keywords,
		KeywordsDef: string(keywordsDefStr),
	}
	id, err := mongodb.ModelWeixinAutoReply.InsertOne(c, &doc)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"id": id})
}

// 删除关注回复、关键词回复、消息回复
func (ctl *AutoReplyController) Delete(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	// 检查当前登录用户是否有管理员权限
	isAdmin, err := ctl.checkIsAdmin(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	if !isAdmin {
		ctl.returnFail(c, 1, "no permission")
		return
	}

	ret, err := mongodb.ModelWeixinAutoReply.DeleteByID(c, id)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{id: id, "result": ret})
}

// 设置是否启用
func (ctl *AutoReplyController) SetEnabled(c *gin.Context) {
	var form struct {
		ReplyType string `json:"reply_type" form:"reply_type" binding:"required"` // subscribe, keyword, message
		Enabled   *bool  `json:"enabled" form:"enabled" binding:"required"`
	}
	if c.ShouldBindJSON(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	flag := lo.Contains([]string{"subscribe", "keyword", "message"}, form.ReplyType)
	if !flag {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	filter := bson.D{{Key: "appid", Value: appid}, {Key: "reply_type", Value: form.ReplyType}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "enabled", Value: form.Enabled}}}}
	ret, err := mongodb.ModelWeixinAutoReply.UpdateMany(ctx, filter, update)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}
	ctl.returnOk(c, gin.H{
		"updated": ret.ModifiedCount,
	})
}
