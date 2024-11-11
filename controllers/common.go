package controllers

import (
	"io"
	"net/http"

	"github.com/anchel/wechat-official-account-admin/routes"
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

// 获取当前服务器的公网IP，调用 https://ipinfo.io/json 接口
func (ctl *CommonController) getPublicIP(c *gin.Context) {
	res, err := http.Get("https://ipinfo.io/ip")
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}
	defer res.Body.Close()
	bs, err := io.ReadAll(res.Body)
	if err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}
	ctl.returnOk(c, gin.H{"ip": string(bs)})
}
