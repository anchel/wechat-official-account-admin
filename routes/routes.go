package routes

import (
	"github.com/gin-gonic/gin"
)

type RouteIntFunc func(*gin.RouterGroup)

var initArr []RouteIntFunc = []RouteIntFunc{}

func AddRouteInitFunc(f RouteIntFunc) {
	initArr = append(initArr, f)
}

func InitRoutes(r *gin.RouterGroup) {
	for _, f := range initArr {
		f(r)
	}
}
