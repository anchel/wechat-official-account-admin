package controllers

import (
	"context"

	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &UserController{
			BaseController: BaseController{},
		}
		r.GET("/system/user/list", ctl.GetUserList)
		r.POST("/system/user/save", ctl.SaveUser)
		r.POST("/system/user/delete", ctl.DeleteUser)

		r.POST("/system/user/change_password", ctl.ChangePassword)
	})
}

type UserController struct {
	BaseController
}

func (ctl *UserController) GetUserList(c *gin.Context) {
	ctx := context.TODO()

	// 检查当前登录用户是否有管理权限
	isAdmin, err := ctl.checkIsAdmin(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	if !isAdmin {
		ctl.returnFail(c, 1, "no permission")
		return
	}

	findOptions := options.Find()
	filter := bson.D{}
	userList, err := mongodb.ModelUser.FindMany(ctx, filter, findOptions)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"list": userList})
}

// 保存用户，如果有id就更新，没有就新增
func (ctl *UserController) SaveUser(c *gin.Context) {
	var form struct {
		ID       string `json:"id" form:"id"`
		UserType string `json:"user_type" form:"user_type"`
		Username string `json:"username" form:"username" binding:"required"`
		Password string `json:"password" form:"password" binding:"required"`
		Email    string `json:"email" form:"email"`
		Remark   string `json:"remark" form:"remark"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	ctx := context.Background()

	// 检查当前登录用户是否有管理权限
	_, userName, isAmin, err := ctl.getCurrentUser(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	if !isAmin {
		ctl.returnFail(c, 1, "no permission")
		return
	}

	if form.ID != "" {
		doc, err := mongodb.ModelUser.FindByID(ctx, form.ID)
		if err != nil {
			ctl.returnFail(c, 1, err.Error())
			return
		}
		if doc == nil {
			ctl.returnFail(c, 1, "user not found")
			return
		}

		if userName != "admin" && doc.Username == "admin" {
			ctl.returnFail(c, 1, "only admin can update admin")
			return
		}

		update := bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "user_type", Value: form.UserType},
				{Key: "username", Value: form.Username},
				{Key: "password", Value: form.Password},
				{Key: "email", Value: form.Email},
				{Key: "remark", Value: form.Remark},
			}},
		}
		ret, err := mongodb.ModelUser.UpdateByID(ctx, form.ID, update)
		if err != nil {
			ctl.returnFail(c, 1, err.Error())
			return
		}
		ctl.returnOk(c, gin.H{"id": form.ID, "result": ret})
		return
	}

	// 查找是否存在相同的username或email
	filter := bson.D{
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "username", Value: form.Username}},
			bson.D{{Key: "email", Value: form.Email}},
		}},
	}

	count, err := mongodb.ModelUser.Count(ctx, filter)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	if count > 0 {
		ctl.returnFail(c, 1, "用户名或邮箱已存在")
		return
	}

	user := &mongodb.EntityUser{
		UserType: form.UserType,
		Username: form.Username,
		Password: form.Password,
		Email:    form.Email,
		Remark:   form.Remark,
	}
	ret, err := mongodb.ModelUser.InsertOne(ctx, user)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"id": ret})
}

// func (ctl *UserController) SoftDeleteUser(c *gin.Context) {
// 	var form struct {
// 		Username string `json:"username" binding:"required"`
// 	}
// 	if err := c.ShouldBindJSON(&form); err != nil {
// 		ctl.returnFail(c, 1, "param error")
// 		return
// 	}

// 	now := time.Now()
// 	ctx := context.TODO()
// 	filter := bson.D{{Key: "username", Value: form.Username}}

// 	update := bson.D{
// 		{Key: "$set", Value: bson.D{
// 			{Key: "deleted_at", Value: now},
// 		}},
// 	}

// 	user, err := mongodb.ModelUser.UpdateOne(ctx, filter, update)
// 	if err != nil {
// 		ctl.returnFail(c, 1, err.Error())
// 		return
// 	}

// 	ctl.returnOk(c, user)
// }

func (ctl *UserController) DeleteUser(c *gin.Context) {
	var form struct {
		ID string `json:"id" form:"id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

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

	ctx := context.TODO()

	doc, err := mongodb.ModelUser.FindByID(ctx, form.ID)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	if doc == nil {
		ctl.returnFail(c, 1, "user not found")
		return
	}

	if doc.Username == "admin" {
		ctl.returnFail(c, 1, "can not delete admin")
		return
	}

	ret, err := mongodb.ModelUser.DeleteByID(ctx, form.ID)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"id": form.ID, "result": ret})
}

// 当前登录用户修改密码
func (ctl *UserController) ChangePassword(c *gin.Context) {
	var form struct {
		OldPassword string `json:"old_password" form:"old_password" binding:"required"`
		NewPassword string `json:"new_password" form:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	ctx := context.TODO()

	userId, _, _, err := ctl.getCurrentUser(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	// 检查旧密码是否正确
	doc, err := mongodb.ModelUser.FindByID(ctx, userId)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	if doc == nil {
		ctl.returnFail(c, 1, "user not found")
		return
	}
	if doc.Password != form.OldPassword {
		ctl.returnFail(c, 1, "旧密码不正确")
		return
	}

	// guest用户不允许修改密码
	if doc.Username == "guest" {
		ctl.returnFail(c, 1, "guest用户不允许修改密码")
		return
	}

	// 更新密码
	update := bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "password", Value: form.NewPassword},
		}},
	}
	ret, err := mongodb.ModelUser.UpdateByID(ctx, userId, update)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{"id": userId, "result": ret})
}
