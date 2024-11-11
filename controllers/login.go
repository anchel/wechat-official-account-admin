package controllers

import (
	"context"
	"encoding/json"
	"log"

	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := NewLoginController()
		r.POST("/system/user/login", ctl.Login)

		r.GET("/system/user/logout", ctl.Logout)

		r.GET("/system/user/userinfo", ctl.GetUserInfo)
	})
}

func NewLoginController() *LoginController {
	return &LoginController{
		BaseController: BaseController{},
	}
}

type LoginController struct {
	BaseController
}

func (ctl *LoginController) GetUserInfo(c *gin.Context) {
	session := sessions.Default(c)

	login, ok := session.Get("login").(int)
	if !ok || login != 1 {
		ctl.returnFail(c, 1, "no login")
		return
	}

	userStr := session.Get("user").(string)
	var user mongodb.EntityUser
	err := json.Unmarshal([]byte(userStr), &user)
	if err != nil {
		log.Println("/api/user/userinfo Unmarshal fail", err)
	}

	ctl.returnOk(c, user)
}

type loginForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (ctl *LoginController) Login(c *gin.Context) {
	var form loginForm
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "invalid form")
		return
	}

	ctx := context.TODO()

	user, err := mongodb.ModelUser.FindOne(ctx, bson.D{{Key: "username", Value: form.Username}, {Key: "password", Value: form.Password}})
	if ctl.checkError(c, err) != nil {
		return
	}

	if user == nil {
		ctl.returnFail(c, 1, "用户名或密码错误")
		return
	}

	session := sessions.Default(c)
	session.Set("login", 1)

	userBS, err := json.Marshal(user)
	if err == nil {
		log.Println("/api/user/login marshal ok", string(userBS))
		session.Set("user", string(userBS))
	} else {
		log.Println("/api/user/login marshal fail", err)
	}

	if err := session.Save(); err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, user)
}

type registerForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email" binding:"required"`
}

func (ctl *LoginController) Register(c *gin.Context) {
	var form registerForm
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "invalid form")
		return
	}

	ctx := context.TODO()

	// 靠索引保证唯一性
	// user, err := mongodb.ModelUser.FindOne(ctx, bson.D{{Key: "username", Value: form.Username}})
	// if ctl.checkError(c, err) != nil {
	// 	return
	// }

	// if user != nil {
	// 	ctl.returnFail(c, 1, "user exists")
	// 	return
	// }

	doc := &mongodb.EntityUser{
		Username: form.Username,
		Password: form.Password,
		Email:    form.Email,
		// CreatedAt: time.Now(),
	}
	id, err := mongodb.ModelUser.InsertOne(ctx, doc)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, bson.M{"id": id})
}

func (ctl *LoginController) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()

	if err := session.Save(); err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, nil)
}
