package controllers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"

	"github.com/anchel/wechat-official-account-admin/lib/types"
	"github.com/anchel/wechat-official-account-admin/lib/util"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type BaseController struct {
}

func (ctl *BaseController) returnOk(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func (ctl *BaseController) returnFail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
	})
}

// func (ctl *BaseController) getParamId(c *gin.Context) (uint, error) {
// 	idstr := c.Param("id")
// 	id, err := strconv.Atoi(idstr)
// 	if err != nil {
// 		ctl.returnFail(c, 1, "id invalid")
// 		return 0, err
// 	}

// 	return uint(id), nil
// }

func (ctl *BaseController) checkError(c *gin.Context, err error) error {
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
	}
	return err
}

// 返回文件
func (ctl *BaseController) returnFile(c *gin.Context, filePath string) {
	wd, _ := util.GetExePwd()
	filePath = filepath.Join(wd, filePath)
	c.File(filePath)
}

// 获取当前登录用户的id
func (ctl *BaseController) getCurrentUser(c *gin.Context) (string, string, bool, error) {
	session := sessions.Default(c)
	userStr := session.Get("user").(string)
	var user mongodb.EntityUser
	err := json.Unmarshal([]byte(userStr), &user)
	if err != nil {
		log.Println("/api/user/userinfo Unmarshal fail", err)
	}
	return user.ID.Hex(), user.Username, user.UserType == "admin", nil
}

func (ctl *BaseController) newContext(c *gin.Context) (context.Context, string, error) {
	appid, ok := c.Get("appid")
	if !ok {
		return nil, "", errors.New("newContext appid not found in gin.Context")
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, types.ContextKey("appid"), appid.(string))
	return ctx, appid.(string), nil
}

func (ctl *BaseController) checkIsAdmin(ctx *gin.Context) (bool, error) {
	_, _, isAdmin, err := ctl.getCurrentUser(ctx)
	if err != nil {
		return false, err
	}
	return isAdmin, nil
}
