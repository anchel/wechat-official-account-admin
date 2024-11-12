package controllers

import (
	"context"
	"log"
	"mime/multipart"
	"path/filepath"
	"time"

	util "github.com/anchel/wechat-official-account-admin/lib/utils"
	"github.com/anchel/wechat-official-account-admin/modules/weixin"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	materialservice "github.com/anchel/wechat-official-account-admin/services/material-service"
	"github.com/anchel/wechat-official-account-admin/wxmp/wxapi"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &MaterialController{
			BaseController: &BaseController{},
		}
		r.GET("/material/source-local", ctl.SourceLocal)
		r.GET("/material/list", ctl.GetList)
		r.POST("/material/upload", ctl.UploadMaterial)
		r.POST("/material/delete", ctl.DeleteMaterial)
	})
}

type MaterialController struct {
	*BaseController
}

type MaterialSourceLocalForm struct {
	MediaCat  string `json:"media_cat" form:"media_cat" binding:"required"`   // 临时素材，永久素材
	MediaType string `json:"media_type" form:"media_type" binding:"required"` // image,voice,video,thumb
	MediaId   string `json:"media_id" form:"media_id" binding:"required"`
}

func (ctl *MaterialController) SourceLocal(c *gin.Context) {
	var form MaterialSourceLocalForm
	if c.ShouldBindQuery(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}
	if form.MediaCat == "" {
		form.MediaCat = "perm"
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	filter := bson.D{
		{Key: "appid", Value: appid},
		{Key: "media_type", Value: form.MediaType},
		{Key: "media_id", Value: form.MediaId},
	}
	doc, err := mongodb.ModelWeixinMaterial.FindOne(ctx, filter)
	if ctl.checkError(c, err) != nil {
		return
	}
	if doc != nil {
		wd, _ := util.GetExePwd()
		fileDstPath := filepath.Join(wd, doc.FilePath)
		ok, err := util.FileExistsAndAccessible(fileDstPath)
		if err == nil {
			if ok {
				ctl.returnFile(c, doc.FilePath) //直接返回文件流
				return
			}
		}
	}

	log.Println("SourceLocal 本地数据库未找到素材或本地文件不存在，去微信侧下载", form.MediaCat, form.MediaType, form.MediaId)

	var data []byte
	var retMap map[string]string

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	if form.MediaCat == "temp" {
		data, retMap, err = wxApiClient.DownloadTempMaterial(ctx, form.MediaId)
	} else {
		data, retMap, err = wxApiClient.DownloadMaterial(ctx, form.MediaType, form.MediaId)
	}

	if ctl.checkError(c, err) != nil {
		return
	}
	ext := retMap["extension"]

	if ext == "" {
		ext = util.GetExtByMediaType(form.MediaType)
	}

	distFilePath, filePath, err := util.GetWxDownloadMediaFilePath("download-", ext, form.MediaId)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = util.SaveFile(distFilePath, data)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = saveMaterialToDatabase(ctx, appid, form.MediaCat, form.MediaType, form.MediaId, filePath, filePath, retMap["url"], retMap["title"], retMap["description"], nil)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnFile(c, filePath) //直接返回文件流
}

type MaterialGetListForm struct {
	MediaCat  string `json:"media_cat" form:"media_cat" binding:"required"`   // temp-临时素材，perm-永久素材
	MediaType string `json:"media_type" form:"media_type" binding:"required"` // image,voice,video,thumb,news
	Offset    int    `json:"offset" form:"offset"`
	Count     int    `json:"count" form:"count"`
}

func (ctl *MaterialController) GetList(c *gin.Context) {
	var form MaterialGetListForm
	if c.ShouldBindQuery(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}
	// log.Println("GetList form", form)
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	list := make([]*struct {
		ID        string     `json:"id"`
		MediaCat  string     `json:"media_cat"`
		MediaType string     `json:"media_type"`
		MediaId   string     `json:"media_id"`
		Name      string     `json:"name"`
		WxUrl     string     `json:"wx_url" bson:"wx_url"`
		ExpiresAt *time.Time `json:"expires_at"`
		Content   struct {
			NewsItem []*wxapi.MessageNewsItem `json:"news_item"`
		} `json:"content,omitempty"`
	}, 0)

	// 永久素材拉取列表，不支持thumb类型
	if form.MediaCat == "temp" || (form.MediaCat == "perm" && form.MediaType == "thumb") {
		retobj, err := materialservice.GetLocalMaterialList(ctx, form.MediaCat, form.MediaType, form.Offset, form.Count)
		if ctl.checkError(c, err) != nil {
			return
		}
		for _, doc := range retobj.List {
			list = append(list, &struct {
				ID        string     `json:"id"`
				MediaCat  string     `json:"media_cat"`
				MediaType string     `json:"media_type"`
				MediaId   string     `json:"media_id"`
				Name      string     `json:"name"`
				WxUrl     string     `json:"wx_url" bson:"wx_url"`
				ExpiresAt *time.Time `json:"expires_at"`
				Content   struct {
					NewsItem []*wxapi.MessageNewsItem `json:"news_item"`
				} `json:"content,omitempty"`
			}{
				ID:        doc.ID.Hex(),
				MediaCat:  doc.MediaCat,
				MediaType: doc.MediaType,
				MediaId:   doc.MediaId,
				WxUrl:     doc.WxUrl,
				ExpiresAt: doc.ExpiresAt,
			})
		}
		ctl.returnOk(c, gin.H{"total": retobj.Total, "list": list})
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	retobj, err := wxApiClient.GetMaterialList(ctx, form.MediaType, form.Offset, form.Count)
	if ctl.checkError(c, err) != nil {
		return
	}
	for _, item := range retobj.Item {
		list = append(list, &struct {
			ID        string     `json:"id"`
			MediaCat  string     `json:"media_cat"`
			MediaType string     `json:"media_type"`
			MediaId   string     `json:"media_id"`
			Name      string     `json:"name"`
			WxUrl     string     `json:"wx_url" bson:"wx_url"`
			ExpiresAt *time.Time `json:"expires_at"`
			Content   struct {
				NewsItem []*wxapi.MessageNewsItem `json:"news_item"`
			} `json:"content,omitempty"`
		}{
			ID:        "",
			MediaCat:  form.MediaCat,
			MediaType: form.MediaType,
			MediaId:   item.MediaId,
			Name:      item.Name,
			WxUrl:     item.URL,
			ExpiresAt: nil,
			Content:   item.Content,
		})
	}

	ctl.returnOk(c, gin.H{"total": retobj.TotalCount, "list": list})
}

type MaterialUploadMaterialForm struct {
	MediaCat    string                `json:"media_cat" form:"media_cat"`                      // 临时素材，永久素材
	MediaType   string                `json:"media_type" form:"media_type" binding:"required"` // image,voice,video,thumb
	File        *multipart.FileHeader `json:"file" form:"file" binding:"required"`
	Title       string                `json:"title" form:"title"`
	Description string                `json:"description" form:"description"`
}

// 上传素材，支持临时素材和永久素材
func (ctl *MaterialController) UploadMaterial(c *gin.Context) {
	var form MaterialUploadMaterialForm
	if c.ShouldBind(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	file := form.File
	dstFilePath, filePath, err := util.GetUploadFilePath(filepath.Ext(file.Filename))
	if ctl.checkError(c, err) != nil {
		return
	}
	log.Println("upload filepath", dstFilePath, filePath)
	if err := c.SaveUploadedFile(file, dstFilePath); err != nil {
		ctl.returnFail(c, 500, err.Error())
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	var ret *wxapi.UploadMaterialResponse
	var expiresAt time.Time

	if form.MediaCat == "" {
		form.MediaCat = "perm"
	}
	if form.MediaCat == "temp" {
		ret, err = wxApiClient.UploadTempMaterial(ctx, form.MediaType, dstFilePath)
		if ctl.checkError(c, err) != nil {
			return
		}
		// 临时素材的过期时间是72小时
		expiresAt = time.Unix(ret.CreatedAt, 0).Add(72 * time.Hour)
	} else {
		ret, err = wxApiClient.UploadMaterial(ctx, form.MediaType, dstFilePath, map[string]string{"title": form.Title, "introduction": form.Description})
	}

	if ctl.checkError(c, err) != nil {
		return
	}

	err = saveMaterialToDatabase(ctx, appid, form.MediaCat, form.MediaType, ret.MediaId, filePath, filePath, ret.Url, form.Title, form.Description, &expiresAt)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, ret)
}

func saveMaterialToDatabase(ctx context.Context, appid string, mediaCat string, mediaType string, mediaId string, filePath string, fileUrlPath string, url string, title string, description string, expiresAt *time.Time) error {
	filter := bson.D{
		{Key: "appid", Value: appid},
		{Key: "media_type", Value: mediaType},
		{Key: "media_id", Value: mediaId},
	}
	fields := bson.D{
		{Key: "media_cat", Value: mediaCat},
		{Key: "file_path", Value: filePath},
		{Key: "file_url_path", Value: fileUrlPath},
		{Key: "wx_url", Value: url},
	}
	if mediaType == "video" {
		fields = append(fields, bson.E{Key: "title", Value: title})
		fields = append(fields, bson.E{Key: "description", Value: description})
	}
	if mediaCat == "temp" {
		fields = append(fields, bson.E{Key: "expires_at", Value: expiresAt})
	}

	update := bson.D{
		{Key: "$set", Value: fields},
	}

	_, err := mongodb.ModelWeixinMaterial.FindOneAndUpdate(ctx, filter, update, true)
	return err
}

// 删除素材
func (ctl *MaterialController) DeleteMaterial(c *gin.Context) {
	var form struct {
		MediaCat string `json:"media_cat" form:"media_cat"` // temp-临时素材，perm-永久素材
		MediaId  string `json:"media_id" form:"media_id" binding:"required"`
	}
	if c.ShouldBindJSON(&form) != nil {
		ctl.returnFail(c, 400, "参数错误")
		return
	}
	if form.MediaCat == "" {
		form.MediaCat = "perm"
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	// 永久素材，需要先删除微信侧的素材
	if form.MediaCat == "perm" {
		log.Println("DeleteMaterial 删除微信侧的素材", form.MediaCat, form.MediaId)
		wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
		if ctl.checkError(c, err) != nil {
			return
		}

		err = wxApiClient.DeleteMaterial(ctx, form.MediaId)
		if ctl.checkError(c, err) != nil {
			return
		}
	}

	// 删除本地数据库的记录
	filter := bson.D{
		{Key: "appid", Value: appid},
		{Key: "media_cat", Value: form.MediaCat},
		{Key: "media_id", Value: form.MediaId},
	}
	_, err = mongodb.ModelWeixinMaterial.DeleteOne(ctx, filter)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, nil)
}
