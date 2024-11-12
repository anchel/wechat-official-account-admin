package controllers

import (
	"github.com/anchel/wechat-official-account-admin/routes"
	commonservice "github.com/anchel/wechat-official-account-admin/services/common-service"
	"github.com/gin-gonic/gin"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &CommonController{
			BaseController: &BaseController{},
		}
		r.GET("/system/common/public-ip", ctl.getPublicIP)
	})
}

type CommonController struct {
	*BaseController
}

func (ctl *CommonController) getPublicIP(c *gin.Context) {
	ip, err := commonservice.GetPublicIP()
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"ip": ip})
}
