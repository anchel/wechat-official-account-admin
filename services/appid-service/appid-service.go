package appidservice

import (
	"context"
	"errors"
	"os"

	"github.com/anchel/wechat-official-account-admin/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func GetAppIDList(ctx context.Context) ([]*mongodb.EntityWxAppid, error) {
	appList := make([]*mongodb.EntityWxAppid, 0)
	wxAppId := os.Getenv("MP_APPID")
	if wxAppId != "" {
		appList = append(appList, &mongodb.EntityWxAppid{
			AppID:          wxAppId,
			Name:           "默认公众号",
			AppSecret:      os.Getenv("MP_APPSECRET"),
			Token:          os.Getenv("MP_TOKEN"),
			EncodingAESKey: os.Getenv("MP_AESKEY"),
		})
	}
	findOptions := options.Find().SetSort(bson.M{"created_at": -1})
	filter := bson.D{}
	docs, err := mongodb.ModelWxAppid.FindMany(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	appList = append(appList, docs...)
	return appList, nil
}

func GetAppIDInfo(ctx context.Context, appid string) (*mongodb.EntityWxAppid, error) {
	var appidInfo *mongodb.EntityWxAppid
	wxAppId := os.Getenv("MP_APPID")
	if wxAppId != "" && wxAppId == appid {
		appidInfo = &mongodb.EntityWxAppid{
			AppID:          wxAppId,
			Name:           "默认公众号",
			AppSecret:      os.Getenv("MP_APPSECRET"),
			Token:          os.Getenv("MP_TOKEN"),
			EncodingAESKey: os.Getenv("MP_AESKEY"),
		}
	} else {
		filter := bson.D{{Key: "appid", Value: appid}}
		doc, err := mongodb.ModelWxAppid.FindOne(context.Background(), filter)
		if err != nil {
			return nil, err
		}
		appidInfo = doc
	}
	return appidInfo, nil
}

type GetAppEnabledDataResp struct {
	AppID                     string `json:"appid"`
	EnabledAutoReplyKeyword   bool   `json:"enabled_auto_reply_keyword"`
	EnabledAutoReplyMessage   bool   `json:"enabled_auto_reply_message"`
	EnabledAutoReplySubscribe bool   `json:"enabled_auto_reply_subscribe"`
}

// 获取公众号的相关功能启用状态
func GetAppEnabledData(ctx context.Context, appid string) (*GetAppEnabledDataResp, error) {
	filter := bson.D{{Key: "appid", Value: appid}}
	doc, err := mongodb.ModelWxAppid.FindOne(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, errors.New("appid not found")
	}
	return &GetAppEnabledDataResp{
		AppID:                     doc.AppID,
		EnabledAutoReplyKeyword:   doc.EnabledAutoReplyKeyword,
		EnabledAutoReplyMessage:   doc.EnabledAutoReplyMessage,
		EnabledAutoReplySubscribe: doc.EnabledAutoReplySubscribe,
	}, nil
}

// 获取公众号的具体某一功能启用状态
func GetAppEnabledDataForReplyType(ctx context.Context, appid string, reply_type string) (bool, error) {
	filter := bson.D{{Key: "appid", Value: appid}}
	doc, err := mongodb.ModelWxAppid.FindOne(context.Background(), filter)
	if err != nil {
		return false, err
	}
	if doc == nil {
		return false, errors.New("appid not found")
	}
	if reply_type == "keyword" {
		return doc.EnabledAutoReplyKeyword, nil
	} else if reply_type == "message" {
		return doc.EnabledAutoReplyMessage, nil
	} else if reply_type == "subscribe" {
		return doc.EnabledAutoReplySubscribe, nil
	}
	return false, errors.New("reply_type error")
}

// 设置公众号的相关功能启用状态
func SetAppEnabledData(ctx context.Context, appid string, reply_type string, enabled bool) error {
	filter := bson.D{{Key: "appid", Value: appid}}
	var d bson.D
	if reply_type == "keyword" {
		d = bson.D{{Key: "enabled_auto_reply_keyword", Value: enabled}}
	} else if reply_type == "message" {
		d = bson.D{{Key: "enabled_auto_reply_message", Value: enabled}}
	} else if reply_type == "subscribe" {
		d = bson.D{{Key: "enabled_auto_reply_subscribe", Value: enabled}}
	} else {
		return errors.New("reply_type error")
	}
	update := bson.D{
		{Key: "$set", Value: d},
	}
	_, err := mongodb.ModelWxAppid.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	return nil
}
