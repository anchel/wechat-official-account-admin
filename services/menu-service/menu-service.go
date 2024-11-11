package menuservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/anchel/wechat-official-account-admin/lib/types"
	"github.com/anchel/wechat-official-account-admin/mongodb"
	weixinservice "github.com/anchel/wechat-official-account-admin/services/weixin-service"
	"github.com/anchel/wechat-official-account-admin/wxmp/wxapi"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// 获取本地的普通菜单
func GetMenuNormal(ctx context.Context, menu_type string, menu_id string) (*mongodb.EntityMenu, []*wxapi.MenuButtonItemApiFormat, error) {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "menu_type", Value: menu_type}, {Key: "menu_id", Value: menu_id}}
	doc, err := mongodb.ModelMenu.FindOne(ctx, filter)
	if err != nil {
		return nil, nil, err
	}

	if doc == nil {
		return doc, nil, nil
	}

	var menuData map[string][]*wxapi.MenuButtonItemApiFormat
	err = json.Unmarshal([]byte(doc.MenuData), &menuData)
	if err != nil {
		return doc, nil, err
	}
	return doc, menuData["button"], nil
}

// 根据ID获取本地的菜单，包含普通菜单和个性化菜单
func GetMenuByID(ctx context.Context, id string) (string, []*wxapi.MenuButtonItemApiFormat, *wxapi.MenuMatchRule, string, error) {
	doc, err := mongodb.ModelMenu.FindByID(ctx, id)
	if err != nil {
		return "", nil, nil, "", err
	}
	if doc == nil {
		return "", nil, nil, "", errors.New("document not found")
	}

	menuData := struct {
		Button    []*wxapi.MenuButtonItemApiFormat `json:"button"`
		MatchRule *wxapi.MenuMatchRule             `json:"matchrule,omitempty"`
	}{}
	err = json.Unmarshal([]byte(doc.MenuData), &menuData)
	if err != nil {
		log.Println("GetMenuByID json.Unmarshal error", err)
		return "", nil, nil, "", err
	}
	return doc.ID.Hex(), menuData.Button, menuData.MatchRule, doc.MenuId, nil
}

type GetMenuConditionalResponse struct {
	ID        string                                  `json:"id"`
	MatchRule *wxapi.MenuMatchRule                    `json:"matchrule,omitempty"`
	Button    []*wxapi.MenuButtonItemApiFormat        `json:"button"`
	AutoReply map[string]*weixinservice.AutoReplyData `json:"autoreply,omitempty"` // 这个是给外层填充用的
	MenuId    string                                  `json:"menuid"`
}

// 获取本地的个性化菜单列表
func GetMenuConditionalList(ctx context.Context, menu_type string) ([]*GetMenuConditionalResponse, error) {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	findOptions := options.Find()
	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "menu_type", Value: menu_type}}
	docs, err := mongodb.ModelMenu.FindMany(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}

	results := make([]*GetMenuConditionalResponse, 0)
	for _, doc := range docs {
		menuData := GetMenuConditionalResponse{ID: doc.ID.Hex(), MenuId: doc.MenuId}
		err = json.Unmarshal([]byte(doc.MenuData), &menuData)
		if err != nil {
			return nil, err
		}
		results = append(results, &menuData)
	}

	return results, nil
}

// 保存个性化菜单
func SaveMenuConditional(ctx context.Context, menu_type string, id string, matchrule *wxapi.MenuMatchRule, buttons []*wxapi.MenuButtonItemApiFormat) (string, error) {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	data, err := json.Marshal(map[string]interface{}{"button": buttons, "matchrule": matchrule})
	if err != nil {
		return "", err
	}

	if id == "" {
		doc := mongodb.ModelMenu.NewEntity()
		doc.AppID = fmt.Sprint(wxAppId)
		doc.MenuType = menu_type
		doc.MenuData = string(data)
		doc.CreatedAt = time.Now()

		id, err = mongodb.ModelMenu.InsertOne(ctx, doc)
		if err != nil {
			return "", err
		}
		return id, err
	}

	// 检查是否越权
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return "", err
	}
	filter := bson.D{{Key: "_id", Value: objectID}, {Key: "appid", Value: wxAppId}}
	doc, err := mongodb.ModelMenu.FindOne(ctx, filter)
	if err != nil {
		return "", err
	}
	if doc == nil {
		return "", errors.New("document not found")
	}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "menu_data", Value: string(data)}}}}
	updateResult, err := mongodb.ModelMenu.UpdateByID(ctx, id, update)
	if err != nil {
		return "", err
	}
	if updateResult.ModifiedCount == 0 {
		log.Println("SaveMenuConditional updateResult.ModifiedCount == 0", id)
		return "", errors.New("SaveMenuConditional updateResult.ModifiedCount == 0")
	}
	return id, err
}

// 删除个性化菜单
func DeleteMenuConditional(ctx context.Context, id string) error {
	_, err := mongodb.ModelMenu.DeleteByID(ctx, id)
	return err
}

// 删除所有个性化菜单
func DeleteAllMenuConditional(ctx context.Context) error {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "menu_type", Value: "conditional"}}
	_, err := mongodb.ModelMenu.DeleteMany(ctx, filter)
	return err
}

// 保存普通菜单
func SaveMenuNormal(ctx context.Context, menu_type string, menu_id string, buttons []*wxapi.MenuButtonItemApiFormat) (*mongodb.EntityMenu, error) {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	data, err := json.Marshal(map[string][]*wxapi.MenuButtonItemApiFormat{"button": buttons})
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "menu_type", Value: menu_type}, {Key: "menu_id", Value: menu_id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "menu_data", Value: string(data)}}}}
	doc, err := mongodb.ModelMenu.FindOneAndUpdate(ctx, filter, update, true)
	return doc, err
}

// 删除普通菜单
func DeleteMenuNormal(ctx context.Context, menu_type string, menu_id string) error {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "menu_type", Value: menu_type}, {Key: "menu_id", Value: menu_id}}
	_, err := mongodb.ModelMenu.DeleteOne(ctx, filter)
	return err
}

// 根据ID获取对应的回复数据
func GetMenuReplyData(ctx context.Context, extId string, draft bool) (map[string]*weixinservice.AutoReplyData, error) {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "reply_type", Value: string(weixinservice.AutoReplyTypeMenuClick)}, {Key: "ext_id", Value: extId}}
	doc, err := mongodb.ModelWeixinAutoReply.FindOne(ctx, filter)
	if err != nil {
		return nil, err
	}

	if doc == nil {
		return nil, nil
	}

	dataStr := doc.ReplyData
	if draft {
		dataStr = doc.DraftData
	}

	if dataStr == "" {
		return nil, nil
	}

	var replyDataMap = make(map[string]*weixinservice.AutoReplyData)
	err = json.Unmarshal([]byte(dataStr), &replyDataMap)
	if err != nil {
		return nil, err
	}

	return replyDataMap, nil
}

// 保存到草稿数据，等发布菜单到微信时，再更新到正式数据
func SaveMenuReplyData(ctx context.Context, extId string, data map[string]*weixinservice.AutoReplyData) error {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	var dataStr string
	var err error
	if data == nil {
		dataStr = ""
	} else {
		bs, err := json.Marshal(data)
		if err != nil {
			return err
		}
		dataStr = string(bs)
	}

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "reply_type", Value: string(weixinservice.AutoReplyTypeMenuClick)}, {Key: "ext_id", Value: extId}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "draft_data", Value: string(dataStr)}}}}
	_, err = mongodb.ModelWeixinAutoReply.FindOneAndUpdate(ctx, filter, update, true)
	return err
}

// 删除回复数据
func DeleteMenuReplyData(ctx context.Context, extId string) error {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "reply_type", Value: string(weixinservice.AutoReplyTypeMenuClick)}, {Key: "ext_id", Value: extId}}
	_, err := mongodb.ModelWeixinAutoReply.DeleteOne(ctx, filter)
	return err
}

// 草稿转为正式数据
func PublishMenuReplyData(ctx context.Context, extId string) error {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "reply_type", Value: string(weixinservice.AutoReplyTypeMenuClick)}, {Key: "ext_id", Value: extId}}
	doc, err := mongodb.ModelWeixinAutoReply.FindOne(ctx, filter)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "reply_data", Value: doc.DraftData}}}}
	_, err = mongodb.ModelWeixinAutoReply.UpdateOne(ctx, filter, update)
	return err
}

// 更新记录的微信侧menuid
func UpdateMenuID(ctx context.Context, id string, menuid string) error {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	// 检查是否越权
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.D{{Key: "_id", Value: objectID}, {Key: "appid", Value: wxAppId}}
	doc, err := mongodb.ModelMenu.FindOne(ctx, filter)
	if err != nil {
		return err
	}
	if doc == nil {
		return errors.New("document not found")
	}

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "menu_id", Value: menuid}}}}
	_, err = mongodb.ModelMenu.UpdateByID(ctx, id, update)
	return err
}

// 删除所有菜单回复数据
func DeleteAllMenuReplyData(ctx context.Context) error {
	wxAppId := ctx.Value(types.ContextKey("appid"))

	filter := bson.D{{Key: "appid", Value: wxAppId}, {Key: "reply_type", Value: string(weixinservice.AutoReplyTypeMenuClick)}}
	_, err := mongodb.ModelWeixinAutoReply.DeleteMany(ctx, filter)
	return err
}
