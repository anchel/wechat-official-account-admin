package controllers

import (
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &RequestLogController{
			BaseController: &BaseController{},
		}
		r.GET("/request-log/list", ctl.List)
	})
}

type RequestLogController struct {
	*BaseController
}

func (ctl *RequestLogController) List(c *gin.Context) {
	var form struct {
		Offset  *int64 `json:"offset" form:"offset" binding:"required"`
		Count   *int64 `json:"count" form:"count" binding:"required"`
		Keyword string `json:"keyword" form:"keyword"`
	}
	if err := c.ShouldBindQuery(&form); err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	findOptions := options.Find()
	findOptions.SetSkip(*form.Offset)
	findOptions.SetLimit(*form.Count)
	findOptions.SetSort(bson.D{{Key: "time", Value: -1}})
	filter := bson.D{
		{Key: "appid", Value: appid},
	}
	if form.Keyword != "" {
		filter = append(filter, bson.E{Key: "$or", Value: bson.A{bson.D{{Key: "path", Value: bson.D{{Key: "$regex", Value: form.Keyword}}}}, bson.D{{Key: "query", Value: bson.D{{Key: "$regex", Value: form.Keyword}}}}}})
	}

	total, err := mongodb.ModelRequestLog.Count(ctx, filter)
	if ctl.checkError(c, err) != nil {
		return
	}

	docs, err := mongodb.ModelRequestLog.FindMany(ctx, filter, findOptions)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, gin.H{"total": total, "list": docs})
}
