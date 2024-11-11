package controllers

import (
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-gonic/gin"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &ExampleController{
			BaseController: &BaseController{},
		}
		r.GET("/example/list", ctl.List)
	})
}

type ExampleController struct {
	*BaseController
}

func (ctl *ExampleController) List(c *gin.Context) {
	ctl.returnOk(c, nil)
}
