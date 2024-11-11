package weixin

import (
	"context"
	"errors"
	"log"

	"github.com/anchel/wechat-official-account-admin/lib/lru"
	appidservice "github.com/anchel/wechat-official-account-admin/services/appid-service"
	weixinservice "github.com/anchel/wechat-official-account-admin/services/weixin-service"
	mpoptions "github.com/anchel/wechat-official-account-admin/wxmp/mp-options"
	"github.com/anchel/wechat-official-account-admin/wxmp/msghandler"
	"github.com/anchel/wechat-official-account-admin/wxmp/wxapi"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
)

var lruMsgHandler *lru.CacheLRU[msghandler.MsgHandler]
var lruWxApiClient *lru.CacheLRU[wxapi.WxApi]

const max_count = 2

func InitWeixin(rdb *redis.Client) error {

	lruMsgHandler = lru.NewCacheLRU[msghandler.MsgHandler](max_count, func(ctx context.Context, appid string) (*msghandler.MsgHandler, error) {
		options, err := GetMpOptions(ctx, appid)
		if err != nil {
			return nil, err
		}
		wxMsgHandler := msghandler.NewMsgHandler(options, wxapi.NewWxApi(options, rdb))
		wxMsgHandler.SetHandler(MsgHandlerFunc)
		return wxMsgHandler, nil
	})

	lruWxApiClient = lru.NewCacheLRU[wxapi.WxApi](max_count, func(ctx context.Context, appid string) (*wxapi.WxApi, error) {
		options, err := GetMpOptions(ctx, appid)
		if err != nil {
			return nil, err
		}
		return wxapi.NewWxApi(options, rdb), nil
	})

	return nil
}

func GetWxApiClient(ctx context.Context, appid string) (*wxapi.WxApi, error) {
	return lruWxApiClient.Get(ctx, appid)
}

func GetMpOptions(ctx context.Context, appid string) (*mpoptions.MpOptions, error) {
	appidInfo, err := appidservice.GetAppIDInfo(ctx, appid)
	if err != nil {
		return nil, err
	}
	if appidInfo == nil {
		return nil, errors.New("appid not found")
	}
	return mpoptions.NewMpOptions(appidInfo.AppID, appidInfo.AppSecret, appidInfo.Token, appidInfo.EncodingAESKey), nil
}

/**
 *  url path: /wxmp/:appid/handler
 *
 */
func Serve(c *gin.Context) {
	var params struct {
		AppID string `uri:"appid" binding:"required"`
	}
	if err := c.ShouldBindUri(&params); err != nil {
		log.Println("wxmp serve error", err.Error())
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}
	ctx := context.Background()
	msgHandler, err := lruMsgHandler.Get(ctx, params.AppID)
	if err != nil {
		log.Println("wxmp get msghandler error", err.Error())
		c.JSON(200, gin.H{"code": 404, "message": err.Error()})
		return
	}
	msgHandler.Serve(c)
}

func MsgHandlerFunc(rc *msghandler.ReplyCtrl, msg msghandler.Message) {
	if msg == nil {
		rc.GetGinContext().String(200, "success")
		return
	}

	msgType := msg.GetMsgType() // text image voice video shortvideo location link event
	log.Println(" MsgHandlerFunc msgType", msgType)

	if msgType == "event" {
		msgEvent := msg.(*msghandler.MessageEvent)
		if msgEvent.Event == "subscribe" || msgEvent.Event == "unsubscribe" { // 关注/取消关注
			err := weixinservice.DoWxUserSubscribe(rc.GetMsgHandler().GetMpOptions().AppId, msgEvent)
			if err != nil {
				rc.GetGinContext().JSON(200, err)
				return
			}
		} else if msgEvent.Event == "SCAN" { // 扫码
			log.Println("event 扫码事件", msgEvent.EventKey)
			// todo
		} else {
			log.Println("event 其他事件", msgEvent.Event, msgEvent.EventKey)
		}
	}

	msgList, err := weixinservice.GetReplyMessages(rc.GetMsgHandler().GetMpOptions().AppId, msg)
	if err != nil {
		rc.GetGinContext().JSON(200, err)
		return
	}

	hasReply := false

	if len(msgList) > 0 {

		// 第一条消息且是支持的类型，就用被动回复的形式。其他情况用主动发送消息的形式
		for idx, msg := range msgList {
			if idx == 0 && lo.Contains([]string{"text", "image", "voice", "video", "music", "news"}, msg.MsgType) {
				ReplyMessage(rc, msg)
				hasReply = true
			} else {
				err := SendMessage(rc, msg)
				if err != nil {
					log.Println("SendMessage error", idx, err)
				}
			}
		}
	}

	if !hasReply {
		rc.GetGinContext().String(200, "success")
	}
}

// 被动回复的方式
func ReplyMessage(rc *msghandler.ReplyCtrl, msg *weixinservice.AutoReplyMessage) {
	switch msg.MsgType {
	case "text":
		rc.ReplyText(msg.Content)
	case "image":
		rc.ReplyImage(msg.MediaId)
	case "voice":
		rc.ReplyVoice(msg.MediaId)
	case "video":
		rc.ReplyVideo(msg.MediaId, msg.Title, msg.Description)
	case "music":
		rc.ReplyMusic(msg.Title, msg.Description, msg.MusicUrl, msg.HQMusicUrl, msg.ThumbMediaId)
	case "news":
		articles := make([]*msghandler.ReplyMessageArticle, len(msg.Articles))
		for i, article := range msg.Articles {
			articles[i] = &msghandler.ReplyMessageArticle{
				Title:       article.Title,
				Description: article.Description,
				PicUrl:      article.PicUrl,
				Url:         article.Url,
			}
		}
		rc.ReplyNews(articles)
	}
}

// 主动发送消息
func SendMessage(rc *msghandler.ReplyCtrl, msg *weixinservice.AutoReplyMessage) error {
	var err error
	switch msg.MsgType {
	case "text":
		err = rc.SendText(msg.Content)
	case "image":
		err = rc.SendImage(msg.MediaId)
	case "voice":
		err = rc.SendVoice(msg.MediaId)
	case "video":
		err = rc.SendVideo(msg.MediaId, "thumb_media_id", msg.Title, msg.Description) // todo
	case "music":
		err = rc.SendMusic(msg.Title, msg.Description, msg.MusicUrl, msg.HQMusicUrl, msg.ThumbMediaId)
	case "news":
		articles := make([]*msghandler.SendMessageArticle, len(msg.Articles))
		for i, article := range msg.Articles {
			articles[i] = &msghandler.SendMessageArticle{
				Title:       article.Title,
				Description: article.Description,
				PicUrl:      article.PicUrl,
				Url:         article.Url,
			}
		}
		err = rc.SendNews(articles)
	case "mpnews":
		err = rc.SendMpNews(msg.MediaId)
	case "mpnewsarticle":
		err = rc.SendMpNewsArticle(msg.ArticleId)
	case "wxcard":
		err = rc.SendWxCard(msg.CardId)
	case "miniprogrampage":
		err = rc.SendMiniProgramPage(msg.Title, msg.AppId, msg.PagePath, msg.ThumbMediaId)
	}
	return err
}
