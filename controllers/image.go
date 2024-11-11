package controllers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/anchel/wechat-official-account-admin/lib/util"
	"github.com/anchel/wechat-official-account-admin/routes"
	"github.com/gin-gonic/gin"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := NewImageController()
		r.POST("/system/image/upload", ctl.Upload)
	})
}

func NewImageController() *ImageController {
	return &ImageController{
		BaseController: &BaseController{},
	}
}

type ImageController struct {
	*BaseController
}

type BindFile struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

func (ctl *ImageController) Upload(c *gin.Context) {
	// var bindFile BindFile

	// // Bind file
	// if err := c.ShouldBind(&bindFile); err != nil {
	// 	c.JSON(http.StatusOK, gin.H{
	// 		"code":    2,
	// 		"message": err.Error(),
	// 	})
	// 	return
	// }

	// file := bindFile.File

	file, err := c.FormFile("file")
	if err != nil {
		ctl.returnFail(c, 1, fmt.Sprintf("get form file error: %v", err))
		return
	}

	dstFilePath, filePath, err := util.GetUploadFilePath(filepath.Ext(file.Filename))
	if ctl.checkError(c, err) != nil {
		return
	}
	log.Println("upload filepath", dstFilePath, filePath)

	if err := c.SaveUploadedFile(file, dstFilePath); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    2,
			"message": err.Error(),
		})
		return
	}

	ctl.returnOk(c, gin.H{
		"imgUrl": util.MakePublicServeUrl(c, filePath),
	})
}
