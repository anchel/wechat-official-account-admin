package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/anchel/wechat-official-account-admin/modules/weixin"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	"github.com/anchel/wechat-official-account-admin/routes"
	menuservice "github.com/anchel/wechat-official-account-admin/services/menu-service"
	weixinservice "github.com/anchel/wechat-official-account-admin/services/weixin-service"
	"github.com/anchel/wechat-official-account-admin/wxmp/wxapi"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func init() {
	routes.AddRouteInitFunc(func(r *gin.RouterGroup) {
		ctl := &MenuController{
			BaseController: &BaseController{},
		}
		r.GET("/menu/get-wx-menu", ctl.GetWxMenu)
		r.GET("/menu/sync-menu-from-wx", ctl.SyncFromWxMenu)
		r.GET("/menu/get-wx-all-menu", ctl.GetWxMenuConfig)

		r.GET("/menu/get", ctl.Get)

		r.POST("/menu/save", ctl.saveNormal)
		r.POST("/menu/delete", ctl.deleteNormal)
		r.POST("/menu/release", ctl.createWeixinMenuNormal)

		r.GET("/menu/get-conditional-list", ctl.getConditionalList)
		r.POST("/menu/save-conditional", ctl.saveConditionalMenu)
		r.POST("/menu/delete-conditional", ctl.deleteConditionalMenu)
		r.POST("/menu/release-conditional", ctl.createWeixinMenuConditional)
	})
}

type MenuController struct {
	*BaseController
}

// GetWxMenu 获取微信菜单
func (ctl *MenuController) GetWxMenu(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	menu, err := wxApiClient.GetMenu()
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, menu)
}

type MapAutoReplyData map[string]*weixinservice.AutoReplyData

func MenuItemFix(keyGenerrator KeyGenerator, keys map[string]any, oldMap MapAutoReplyData, newMap MapAutoReplyData, item *wxapi.MenuButtonItemApiFormat) *wxapi.MenuButtonItemApiFormat {
	fixedItem := wxapi.MenuButtonItemApiFormat{
		Type:      item.Type,
		Name:      item.Name,
		SubButton: item.SubButton,
	}

	if item.Key != "" {
		keys[item.Key] = true
	}
	log.Println("item.Type, item.Key", item.Type, item.Key)

	if item.Type == "view" {
		fixedItem.Type = "view"
		fixedItem.Url = item.Url
	} else if item.Type == "text" || item.Type == "img" || item.Type == "voice" || item.Type == "video" || item.Type == "news" {
		fixedItem.Type = "click"
		fixedItem.Key = item.Key // 一般这里是空的
		if fixedItem.Key == "" {
			fixedItem.Key = keyGenerrator()

			rd := MenuItemToReplyData(item)
			if rd != nil {
				newMap[fixedItem.Key] = rd
			}
		} else {
			if _, ok := oldMap[fixedItem.Key]; ok {
				newMap[fixedItem.Key] = oldMap[fixedItem.Key]
			} else {
				rd := MenuItemToReplyData(item)
				if rd != nil {
					newMap[fixedItem.Key] = rd
				}
			}
		}
	} else if item.Type == "click" {
		fixedItem.Type = "click"
		fixedItem.Key = item.Key
		if fixedItem.Key == "" {
			fixedItem.Key = keyGenerrator()
			rd := MenuItemToReplyData(item)
			if rd != nil {
				newMap[fixedItem.Key] = rd
			}
		} else {
			if _, ok := oldMap[fixedItem.Key]; ok {
				newMap[fixedItem.Key] = oldMap[fixedItem.Key]
			} else {
				rd := MenuItemToReplyData(item)
				if rd != nil {
					newMap[fixedItem.Key] = rd
				}
			}
		}
	} else if item.Type == "miniprogram" {
		fixedItem.Type = item.Type
		fixedItem.Key = item.Key
		fixedItem.Url = item.Url
		fixedItem.AppId = item.AppId
		fixedItem.PagePath = item.PagePath
	} else { // scancode_push scancode_waitmsg pic_sysphoto pic_photo_or_album pic_weixin location_select media_id article_id article_view_limited
		fixedItem.Type = item.Type
		fixedItem.Key = item.Key

		if fixedItem.Key != "" {
			if _, ok := oldMap[fixedItem.Key]; ok {
				newMap[fixedItem.Key] = oldMap[fixedItem.Key]
			}
		}
		// 因为暂时不支持其他类型，所以下面先注释了，后面可以根据情况，针对某一些类型进行处理
		// if fixedItem.Key == "" {
		// 	fixedItem.Key = keyGenerrator()
		// }
	}
	return &fixedItem
}

func MenuItemToReplyData(item *wxapi.MenuButtonItemApiFormat) *weixinservice.AutoReplyData {
	msg := weixinservice.AutoReplyMessage{}

	if item.Type == "text" {
		msg.MsgType = "text"
		msg.Content = item.Value
	} else if item.Type == "img" {
		msg.MsgType = "image"
		msg.MediaId = item.Value
	} else if item.Type == "voice" {
		msg.MsgType = "voice"
		msg.MediaId = item.Value
	} else if item.Type == "video" {
		msg.MsgType = "video"
		msg.MediaId = item.Value // todo 这里的value是视频下载地址，怎么获取media_id呢
		msg.Title = "视频标题"
		msg.Description = "视频描述"
	} else if item.Type == "news" {
		msg.MsgType = "news"
		msg.MediaId = item.Value
		var articles []*weixinservice.AutoReplyMessageArticle
		for _, newsItem := range item.NewsInfo.List {
			articles = append(articles, &weixinservice.AutoReplyMessageArticle{
				Title:       newsItem.Title,
				Description: newsItem.Digest,
				PicUrl:      newsItem.CoverUrl,
				Url:         newsItem.ContentUrl,
			})
		}
		msg.Articles = articles
	} else {
		log.Println("你的配置回复类型属于其他:", item.Type)
		return nil
	}

	msgWrapper := weixinservice.AutoReplyData{
		ReplyAll: true,
		MsgList:  []*weixinservice.AutoReplyMessage{&msg},
	}

	// bytes, err := json.Marshal(msgWrapper)
	// if err != nil {
	// 	return ""
	// }
	// return string(bytes)
	return &msgWrapper
}

type KeyGenerator func() string

func CreateKeyGenerator(prefix string, start int32, keys map[string]any) KeyGenerator {
	return func() string {
		var key string
		for {
			start += 1
			key = fmt.Sprintf("%s%d", prefix, start)
			if _, ok := keys[key]; !ok {
				keys[key] = true
				break
			}
		}
		return key
	}
}

func CollectAllKeys(keys map[string]any, buttons []*wxapi.MenuButtonItem) {
	for _, button := range buttons {
		if button.Key != "" {
			keys[button.Key] = true
		}
		if button.SubButton != nil && len(button.SubButton.List) > 0 {
			CollectAllKeys(keys, button.SubButton.List)
		}
	}
}

// 同步微信普通菜单到本地
func (ctl *MenuController) SyncFromWxMenu(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	replyDataMap, err := menuservice.GetMenuReplyData(ctx, "normal", false)
	if ctl.checkError(c, err) != nil {
		return
	}

	var newReplyDataMap = make(map[string]*weixinservice.AutoReplyData)

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	menu, err := wxApiClient.GetMenu()
	if ctl.checkError(c, err) != nil {
		return
	}

	keys := make(map[string]any, 0)
	keyGenerrator := CreateKeyGenerator("key_", 0, keys)

	CollectAllKeys(keys, menu.SelfMenu.Button) // 先收集所有的key

	buttons := make([]*wxapi.MenuButtonItemApiFormat, 0)
	for _, button := range menu.SelfMenu.Button {
		var subButtons []*wxapi.MenuButtonItemApiFormat
		if button.SubButton != nil && len(button.SubButton.List) > 0 {
			for _, subButton := range button.SubButton.List {
				subButtons = append(subButtons, MenuItemFix(keyGenerrator, keys, replyDataMap, newReplyDataMap, &wxapi.MenuButtonItemApiFormat{
					Name:     subButton.Name,
					Type:     subButton.Type,
					Key:      subButton.Key,
					Value:    subButton.Value,
					Url:      subButton.Url,
					NewsInfo: subButton.NewsInfo,
				}))
			}
		}
		buttons = append(buttons, MenuItemFix(keyGenerrator, keys, replyDataMap, newReplyDataMap, &wxapi.MenuButtonItemApiFormat{
			Name:      button.Name,
			Type:      button.Type,
			Key:       button.Key,
			Value:     button.Value,
			Url:       button.Url,
			NewsInfo:  button.NewsInfo,
			SubButton: subButtons,
		}))
	}

	log.Println("newReplyData", newReplyDataMap)

	// 调用微信接口创建菜单
	_, err = wxApiClient.CreateMenu(buttons)
	if ctl.checkError(c, err) != nil {
		return
	}

	// 更新 menus 表
	btnStr, err := json.Marshal(map[string]interface{}{"button": buttons})
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	update2 := bson.D{{Key: "$set", Value: bson.D{{Key: "menu_data", Value: string(btnStr)}}}}
	filter2 := bson.D{{Key: "appid", Value: appid}, {Key: "menu_type", Value: "normal"}, {Key: "menu_id", Value: "normal"}}
	_, err = mongodb.ModelMenu.FindOneAndUpdate(ctx, filter2, update2, true)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	// 更新 autoreply 表
	nrdJsonStr, err := json.Marshal(newReplyDataMap)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}
	filter := bson.D{{Key: "appid", Value: appid}, {Key: "reply_type", Value: string(weixinservice.AutoReplyTypeMenuClick)}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "reply_data", Value: string(nrdJsonStr)}, {Key: "draft_data", Value: string(nrdJsonStr)}}}}
	_, err = mongodb.ModelWeixinAutoReply.FindOneAndUpdate(ctx, filter, update, true)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	ctl.returnOk(c, gin.H{
		"button":    buttons,
		"autoreply": newReplyDataMap,
	})
}

type MenuGetForm struct {
	ID       string `json:"id" form:"id"`
	MenuType string `json:"menu_type" form:"menu_type"`
	MenuId   string `json:"menu_id" form:"menu_id"`
}

/**
 * 拉取本地数据库的菜单数据
 */
func (ctl *MenuController) Get(c *gin.Context) {
	var form MenuGetForm
	var err error

	if err = c.ShouldBindQuery(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	ctx, _, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	var ID string
	var buttons []*wxapi.MenuButtonItemApiFormat
	var matchrule *wxapi.MenuMatchRule
	var menuId string

	if form.ID != "" {
		log.Println("GetMenuByID", form.ID)
		id, bs, mr, mid, err := menuservice.GetMenuByID(ctx, form.ID)
		if ctl.checkError(c, err) != nil {
			return
		}
		ID = id
		menuId = mid
		buttons = bs
		matchrule = mr
	} else {
		doc, bs, err := menuservice.GetMenuNormal(ctx, form.MenuType, form.MenuId)
		if ctl.checkError(c, err) != nil {
			return
		}
		if doc != nil {
			ID = doc.ID.Hex()
			buttons = bs
		}
	}

	var replydata map[string]*weixinservice.AutoReplyData
	if ID != "" {
		replydata, err = menuservice.GetMenuReplyData(ctx, ID, true) // 本地操作的时候，拉取的是草稿数据
		if ctl.checkError(c, err) != nil {
			return
		}
	}

	ctl.returnOk(c, gin.H{
		"id":        ID,
		"button":    buttons,
		"autoreply": replydata,
		"matchrule": matchrule,
		"menuid":    menuId,
	})
}

/**
 * 拉取本地的个性化菜单数据
 */
func (ctl *MenuController) getConditionalList(c *gin.Context) {
	ctx, _, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	list, err := menuservice.GetMenuConditionalList(ctx, "conditional")
	if ctl.checkError(c, err) != nil {
		return
	}

	for _, item := range list {
		replydata, err := menuservice.GetMenuReplyData(ctx, item.ID, true) // 本地操作的时候，拉取的是草稿数据
		if ctl.checkError(c, err) != nil {
			return
		}
		item.AutoReply = replydata
	}

	ctl.returnOk(c, gin.H{
		"list": list,
	})
}

type MenuSaveNormalForm struct {
	Button    []*wxapi.MenuButtonItemApiFormat        `json:"button"`
	AutoReply map[string]*weixinservice.AutoReplyData `json:"autoreply,omitempty"`
}

/*
 * 在本地数据库保存普通菜单
 */
func (ctl *MenuController) saveNormal(c *gin.Context) {
	var form MenuSaveNormalForm
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	ctx, _, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	buttons := make([]*wxapi.MenuButtonItemApiFormat, 0)
	for _, button := range form.Button {
		var subButtons []*wxapi.MenuButtonItemApiFormat
		if button != nil && button.SubButton != nil && len(button.SubButton) > 0 {
			for _, subButton := range button.SubButton {
				subButtons = append(subButtons, &wxapi.MenuButtonItemApiFormat{
					Name:     subButton.Name,
					Type:     subButton.Type,
					Key:      subButton.Key,
					Url:      subButton.Url,
					AppId:    subButton.AppId,
					PagePath: subButton.PagePath,
				})
			}
		}
		buttons = append(buttons, &wxapi.MenuButtonItemApiFormat{
			Name:      button.Name,
			Type:      button.Type,
			Key:       button.Key,
			Url:       button.Url,
			AppId:     button.AppId,
			PagePath:  button.PagePath,
			SubButton: subButtons,
		})
	}

	doc, err := menuservice.SaveMenuNormal(ctx, "normal", "normal", buttons)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = menuservice.SaveMenuReplyData(ctx, doc.ID.Hex(), form.AutoReply)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, gin.H{
		"id": doc.ID.Hex(),
	})
}

type MenuCreateWeixinMenuForm struct {
	MenuType string `json:"menu_type" form:"menu_type" binding:"required"`
	MenuId   string `json:"menu_id" form:"menu_id" binding:"required"`
}

// 在本地数据库删除普通菜单。注意：删除普通菜单，会同时删除所有的个性化菜单
func (ctl *MenuController) deleteNormal(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	// 删除微信侧的菜单
	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	err = wxApiClient.DeleteMenu(ctx)
	if ctl.checkError(c, err) != nil {
		return
	}

	// 删除本地数据库的普通菜单
	err = menuservice.DeleteMenuNormal(ctx, "normal", "normal")
	if ctl.checkError(c, err) != nil {
		return
	}

	// 删除本地数据库的个性化菜单
	err = menuservice.DeleteAllMenuConditional(ctx)
	if ctl.checkError(c, err) != nil {
		return
	}

	// 删除本地数据库的自动回复数据
	err = menuservice.DeleteAllMenuReplyData(ctx)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, nil)
}

/**
 * 调用微信接口创建普通菜单
 */
func (ctl *MenuController) createWeixinMenuNormal(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	doc, buttons, err := menuservice.GetMenuNormal(ctx, "normal", "normal")
	if ctl.checkError(c, err) != nil {
		return
	}
	if doc == nil {
		ctl.returnFail(c, 1, "menu not exists")
		return
	}

	// log.Println("buttons", buttons)

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	_, err = wxApiClient.CreateMenu(buttons)
	if ctl.checkError(c, err) != nil {
		return
	}

	id := doc.ID.Hex()
	err = menuservice.PublishMenuReplyData(ctx, id)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, id)
}

type MenuSaveConditionalMenuForm struct {
	ID        string                                  `json:"id" form:"id"`
	MatchRule *wxapi.MenuMatchRule                    `json:"matchrule,omitempty"`
	Button    []*wxapi.MenuButtonItemApiFormat        `json:"button"`
	AutoReply map[string]*weixinservice.AutoReplyData `json:"autoreply,omitempty"`
}

// 在本地数据库保存个性化菜单
func (ctl *MenuController) saveConditionalMenu(c *gin.Context) {
	var form MenuSaveConditionalMenuForm
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error"+err.Error())
		return
	}

	ctx, _, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	buttons := make([]*wxapi.MenuButtonItemApiFormat, 0)
	for _, button := range form.Button {
		var subButtons []*wxapi.MenuButtonItemApiFormat
		if button != nil && button.SubButton != nil && len(button.SubButton) > 0 {
			for _, subButton := range button.SubButton {
				subButtons = append(subButtons, &wxapi.MenuButtonItemApiFormat{
					Name:     subButton.Name,
					Type:     subButton.Type,
					Key:      subButton.Key,
					Url:      subButton.Url,
					AppId:    subButton.AppId,
					PagePath: subButton.PagePath,
				})
			}
		}
		buttons = append(buttons, &wxapi.MenuButtonItemApiFormat{
			Name:      button.Name,
			Type:      button.Type,
			Key:       button.Key,
			Url:       button.Url,
			AppId:     button.AppId,
			PagePath:  button.PagePath,
			SubButton: subButtons,
		})
	}

	id, err := menuservice.SaveMenuConditional(ctx, "conditional", form.ID, form.MatchRule, buttons)
	if ctl.checkError(c, err) != nil {
		return
	}

	autoReply := form.AutoReply
	keyMap := make(map[string]string)

	if form.ID == "" {
		// 更新每个button的key
		for _, button := range buttons {
			if button != nil {
				newKey := strings.Replace(button.Key, "_new_", fmt.Sprint("_", id, "_"), 1)
				keyMap[button.Key] = newKey
				button.Key = newKey

				if len(button.SubButton) > 0 {
					for _, subButton := range button.SubButton {
						newKey := strings.Replace(subButton.Key, "_new_", fmt.Sprint("_", id, "_"), 1)
						keyMap[subButton.Key] = newKey
						subButton.Key = newKey
					}
				}
			}
		}

		// 重新保存到数据库
		log.Println("第二次保存到数据库", id)
		_, err := menuservice.SaveMenuConditional(ctx, "conditional", id, form.MatchRule, buttons)
		if ctl.checkError(c, err) != nil {
			return
		}

		// 更新自动回复数据中的key
		newAutoReply := make(map[string]*weixinservice.AutoReplyData)
		for key, data := range autoReply {
			newKey, ok := keyMap[key]
			if !ok {
				ctl.returnFail(c, 1, "key not found:"+key)
				return
			}
			newAutoReply[newKey] = data
		}
		autoReply = newAutoReply
	}

	// 保存自动回复数据
	err = menuservice.SaveMenuReplyData(ctx, id, autoReply)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, gin.H{
		"id": id,
	})
}

type MenuPostIdForm struct {
	ID string `json:"id" form:"id" binding:"required"`
}

// 删除个性化菜单
func (ctl *MenuController) deleteConditionalMenu(c *gin.Context) {
	var form MenuPostIdForm
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	// 先检查是否已经在微信侧创建了菜单
	_, _, _, menuId, err := menuservice.GetMenuByID(ctx, form.ID)
	if ctl.checkError(c, err) != nil {
		return
	}

	if menuId != "" {
		// 先在微信端删除个性化菜单
		wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
		if ctl.checkError(c, err) != nil {
			return
		}

		err = wxApiClient.DeleteMenuConditional(menuId)
		if ctl.checkError(c, err) != nil {
			return
		}
	}

	// 在本地数据库删除
	err = menuservice.DeleteMenuConditional(ctx, form.ID)
	if ctl.checkError(c, err) != nil {
		return
	}

	// 删除本地数据库的自动回复数据
	err = menuservice.DeleteMenuReplyData(ctx, form.ID)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, nil)
}

type MenuCreateWeixinMenuConditionalForm struct {
	ID string `json:"id" form:"id" binding:"required"`
}

// 调用微信接口创建个性化菜单
func (ctl *MenuController) createWeixinMenuConditional(c *gin.Context) {
	var form MenuCreateWeixinMenuConditionalForm
	if err := c.ShouldBindJSON(&form); err != nil {
		ctl.returnFail(c, 1, "param error")
		return
	}

	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	id, buttons, matchrule, menuId, err := menuservice.GetMenuByID(ctx, form.ID)
	if ctl.checkError(c, err) != nil {
		return
	}

	if id == "" {
		ctl.returnFail(c, 1, "menu not exists")
		return
	}

	if menuId == "" {
		// todo 需要检查是否创建了普通菜单，只有创建了普通菜单才能创建个性化菜单

		// 调用微信接口创建菜单
		wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
		if ctl.checkError(c, err) != nil {
			return
		}

		menuid, err := wxApiClient.CreateMenuConditional(buttons, matchrule)
		if ctl.checkError(c, err) != nil {
			return
		}

		// 更新menuid到记录中
		err = menuservice.UpdateMenuID(ctx, form.ID, menuid)
		if ctl.checkError(c, err) != nil {
			return
		}

		menuId = menuid
	}

	// 发布草稿数据
	err = menuservice.PublishMenuReplyData(ctx, form.ID)
	if ctl.checkError(c, err) != nil {
		return
	}

	ctl.returnOk(c, gin.H{
		"id":     id,
		"menuid": menuId,
	})
}

// 获取微信侧的所有菜单配置
func (ctl *MenuController) GetWxMenuConfig(c *gin.Context) {
	ctx, appid, err := ctl.newContext(c)
	if err != nil {
		ctl.returnFail(c, 1, err.Error())
		return
	}

	wxApiClient, err := weixin.GetWxApiClient(ctx, appid)
	if ctl.checkError(c, err) != nil {
		return
	}

	menu, err := wxApiClient.GetAllMenu()
	if ctl.checkError(c, err) != nil {
		return
	}
	ctl.returnOk(c, menu)
}
