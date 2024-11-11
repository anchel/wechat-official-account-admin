package weixinservice

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"math/rand"

	"github.com/anchel/wechat-official-account-admin/mongodb"
	appidservice "github.com/anchel/wechat-official-account-admin/services/appid-service"
	"github.com/anchel/wechat-official-account-admin/wxmp/msghandler"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DoWxUserSubscribe(appid string, msg *msghandler.MessageEvent) error {
	fmt.Println("DoWxUserSubscribe", msg)

	filter := bson.D{{Key: "appid", Value: appid}, {Key: "openid", Value: msg.FromUserName}}

	subscribed := false
	if msg.Event == "subscribe" {
		subscribed = true
	}

	// log.Println("msg.EventKey", msg.EventKey)

	update := bson.D{{Key: "$set", Value: bson.D{{Key: "subscribed", Value: subscribed}}}}

	if msg.Event == "subscribe" {
		SceneID, ok := strings.CutPrefix(msg.EventKey, "qrscene_")
		if ok {
			log.Println("SceneID", SceneID)
			update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "scene_id", Value: SceneID}}})
		}
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "subscribed_at", Value: time.Now()}}})
	} else if msg.Event == "unsubscribe" {
		update = append(update, bson.E{Key: "$set", Value: bson.D{{Key: "unsubscribed_at", Value: time.Now()}}})
	}

	ctx := context.Background()

	_, err := mongodb.ModelWeixinUser.FindOneAndUpdate(ctx, filter, update, true)
	if err != nil {
		return err
	}
	return nil
}

type AutoReplyType string

const (
	AutoReplyTypeSubscribe AutoReplyType = "subscribe"
	AutoReplyTypeMessage   AutoReplyType = "message"
	AutoReplyTypeKeyword   AutoReplyType = "keyword"
	AutoReplyTypeMenuClick AutoReplyType = "menu_click"
)

type AutoReplyMessageArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	PicUrl      string `json:"pic_url"`
	Url         string `json:"url"`
}

type AutoReplyMessage struct {
	MsgType      string                     `json:"msg_type"` // text image voice video music news mpnews mpnewsarticle wxcard miniprogrampage
	Content      string                     `json:"content"`
	MediaId      string                     `json:"media_id,omitempty"`
	Title        string                     `json:"title,omitempty"`
	Description  string                     `json:"description,omitempty"`
	ThumbMediaId string                     `json:"thumb_media_id,omitempty"`
	MusicUrl     string                     `json:"music_url,omitempty"`
	HQMusicUrl   string                     `json:"hq_music_url,omitempty"`
	Articles     []*AutoReplyMessageArticle `json:"articles,omitempty"`
	ArticleId    string                     `json:"article_id,omitempty"`
	AppId        string                     `json:"appid,omitempty"`
	PagePath     string                     `json:"pagepath,omitempty"`
	CardId       string                     `json:"card_id,omitempty"`
}

type AutoReplyData struct {
	ReplyAll bool                `json:"reply_all"`
	MsgList  []*AutoReplyMessage `json:"msg_list"`
}

func GetReplyMessages(appid string, msg msghandler.Message) ([]*AutoReplyMessage, error) {
	msgType := msg.GetMsgType() // text image voice video shortvideo location link event

	var replyType AutoReplyType

	if msgType == "event" {
		msg := msg.(*msghandler.MessageEvent)
		if msg.Event == "subscribe" {
			replyType = AutoReplyTypeSubscribe
		} else if msg.Event == "CLICK" {
			replyType = AutoReplyTypeMenuClick
		}
		// 其他的 location_select pic_sysphoto pic_photo_or_album pic_weixin scancode_push scancode_waitmsg pic_weixin 也暂时不处理吧
	} else if msgType == "text" {
		replyType = AutoReplyTypeKeyword
	} else if msgType == "image" || msgType == "voice" || msgType == "video" {
		replyType = AutoReplyTypeMessage
	}
	// 其他的暂时不处理吧 shortvideo location link

	if replyType == "" {
		return nil, nil
	}

	var msgList []*AutoReplyMessage
	var err error
	if replyType == AutoReplyTypeMenuClick {
		msgList, err = GetReplyMessagesForMenuClick(appid, replyType, msg.(*msghandler.MessageEvent).EventKey)
	} else if replyType == AutoReplyTypeKeyword {
		msgList, err = GetReplyMessagesForKeyword(appid, replyType, msg.(*msghandler.MessageText).Content)
		if err != nil {
			return msgList, err
		}
		if len(msgList) <= 0 {
			log.Println("关键词回复为空，改为普通消息回复")
			replyType = AutoReplyTypeMessage // 改变获取类型
			msgList, err = GetReplyMessagesForCommon(appid, replyType, msg)
		}
	} else {
		msgList, err = GetReplyMessagesForCommon(appid, replyType, msg)
	}
	return msgList, err
}

// 菜单点击回复
func GetReplyMessagesForMenuClick(appid string, replyType AutoReplyType, key string) ([]*AutoReplyMessage, error) {
	log.Println("GetReplyMessagesForMenuClick", appid, replyType, key)
	findOptions := options.Find()
	filter := bson.D{{Key: "appid", Value: appid}, {Key: "reply_type", Value: string(replyType)}}
	docs, err := mongodb.ModelWeixinAutoReply.FindMany(context.Background(), filter, findOptions)
	if err != nil {
		log.Println("Error GetReplyMessagesForMenuClick", err)
		return nil, err
	}
	if len(docs) <= 0 {
		log.Println("Error GetReplyMessagesForMenuClick", "docs is nil")
		return nil, nil
	}

	for _, doc := range docs {
		if doc.ReplyData == "" {
			log.Println("Error GetReplyMessagesForMenuClick", "reply_data is empty", replyType, doc.ExtId)
			continue
		}
		// log.Println("GetReplyMessagesForMenuClick", "doc", doc)
		var replyDataMap = make(map[string]*AutoReplyData)
		err = json.Unmarshal([]byte(doc.ReplyData), &replyDataMap)
		if err != nil {
			log.Println("Error json.Unmarshal", err)
			return nil, err
		}
		rd, ok := replyDataMap[key]
		if ok {
			log.Println("GetReplyMessagesForMenuClick 找到key", key)
			return ConvertReplyDataToMessages(rd)
		}
	}

	return nil, nil
}

type KeywordDef struct {
	Keyword string `json:"keyword"`
	Exact   bool   `json:"exact"`
}

/**
 * 关键词回复
 * 先查出全部的规则，然后在进行匹配
 * @param keyword 关键词
 */
func GetReplyMessagesForKeyword(appid string, replyType AutoReplyType, keyword string) ([]*AutoReplyMessage, error) {
	log.Println("GetReplyMessagesForKeyword", appid, replyType, keyword)

	// 检查回复的开关是否已经打开
	enabled, err := appidservice.GetAppEnabledDataForReplyType(context.Background(), appid, string(replyType))
	if err != nil {
		log.Println("Error GetAppEnabledDataForReplyType", err)
		return nil, err
	}
	if !enabled {
		log.Println("GetReplyMessagesForKeyword", replyType, "reply is disabled")
		return nil, nil
	}

	findOptions := options.Find()
	filter := bson.D{{Key: "appid", Value: appid}, {Key: "reply_type", Value: string(replyType)}}
	docs, err := mongodb.ModelWeixinAutoReply.FindMany(context.Background(), filter, findOptions)
	if err != nil {
		log.Println("Error GetReplyMessagesForKeyword", err)
		return nil, err
	}
	if len(docs) <= 0 {
		log.Println("Error GetReplyMessagesForKeyword", "docs is nil")
		return nil, nil
	}
	log.Println("GetReplyMessagesForKeyword", "docs", len(docs))

	for _, doc := range docs {
		if doc.KeywordsDef == "" || doc.ReplyData == "" {
			continue
		}
		keywordsDef := []KeywordDef{}
		err = json.Unmarshal([]byte(doc.KeywordsDef), &keywordsDef)
		if err != nil {
			log.Println("Error json.Unmarshal", doc.RuleTitle, err)
			continue
		}

		match := false
		for _, def := range keywordsDef {
			if def.Exact {
				if def.Keyword == keyword {
					match = true
					break
				}
			} else {
				if strings.Contains(keyword, def.Keyword) {
					match = true
					break
				}
			}
		}

		// 只取匹配到的第一个规则
		if match {
			log.Println("GetReplyMessagesForKeyword", "matched", doc.RuleTitle)
			return ConvertReplyDataToMessages(doc.ReplyData)
		}
	}

	return nil, nil
}

// 订阅和消息回复，都属于公共的
func GetReplyMessagesForCommon(appid string, replyType AutoReplyType, msg msghandler.Message) ([]*AutoReplyMessage, error) {
	log.Println("GetReplyMessagesForCommon", appid, replyType)

	// 检查回复的开关是否已经打开
	enabled, err := appidservice.GetAppEnabledDataForReplyType(context.Background(), appid, string(replyType))
	if err != nil {
		log.Println("Error GetAppEnabledDataForReplyType", err)
		return nil, err
	}
	if !enabled {
		log.Println("GetReplyMessagesForKeyword", replyType, "reply is disabled")
		return nil, nil
	}

	filter := bson.D{{Key: "appid", Value: appid}, {Key: "reply_type", Value: string(replyType)}}
	doc, err := mongodb.ModelWeixinAutoReply.FindOne(context.Background(), filter)
	if err != nil {
		log.Println("Error ModelWeixinAutoReply.FindOne", err)
		return nil, err
	}
	if doc == nil || doc.ReplyData == "" {
		log.Println("Error ModelWeixinAutoReply.FindOne", "doc is nil or reply_data is empty", doc)
		return nil, nil
	}

	return ConvertReplyDataToMessages(doc.ReplyData)
}

/**
 * 将回复数据转换为消息列表
 * @param replyData { replay_all: true, msg_list: [{}, {}] }
 */
func ConvertReplyDataToMessages(replyData any) ([]*AutoReplyMessage, error) {
	data := AutoReplyData{}

	if str, ok := replyData.(string); ok {
		err := json.Unmarshal([]byte(str), &data)
		if err != nil {
			log.Println("Error json.Unmarshal", err)
			return nil, err
		}
	}

	if d, ok := replyData.(*AutoReplyData); ok {
		data = *d
	}

	// log.Println("ConvertReplyDataToMessages", data)

	if len(data.MsgList) <= 0 {
		log.Println("Error ConvertReplyDataToMessages", "msg_list is empty")
		return nil, nil
	}

	if data.ReplyAll {
		log.Println("ConvertReplyDataToMessages", "reply all", len(data.MsgList))
		return data.MsgList, nil
	}

	// 生成一个 0 到 9 的随机整数（不包括 10）
	randomNumber := rand.Intn(len(data.MsgList))
	log.Println("ConvertReplyDataToMessages", "randomNumber", randomNumber)
	return []*AutoReplyMessage{data.MsgList[randomNumber]}, nil
}
