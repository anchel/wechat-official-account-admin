package msghandler

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/anchel/wechat-official-account-admin/wxmp/common"
	mpoptions "github.com/anchel/wechat-official-account-admin/wxmp/mp-options"
	"github.com/anchel/wechat-official-account-admin/wxmp/wxapi"
	"github.com/gin-gonic/gin"
)

type Message interface {
	GetMsgType() string
	GetToUserName() string
	GetFromUserName() string
}

type MsgHandlerFunc func(rc *ReplyCtrl, msg Message)

type ReplyCtrl struct {
	c          *gin.Context
	msgHandler *MsgHandler

	msg Message
}

func (rc *ReplyCtrl) GetGinContext() *gin.Context {
	return rc.c
}

func (rc *ReplyCtrl) GetMsgHandler() *MsgHandler {
	return rc.msgHandler
}

func (rc *ReplyCtrl) GetWxApiClient() *wxapi.WxApi {
	return rc.msgHandler.GetWxApiClient()
}

func (rc *ReplyCtrl) GetMsg() Message {
	return rc.msg
}

func (rc *ReplyCtrl) GetEncryptBytes(rawBytes []byte) ([]byte, error) {
	encryptedMsg, err := common.AesEncrypt(rawBytes, rc.msgHandler.mpoptions.AesKey)
	if err != nil {
		return nil, err
	}

	timestamp := fmt.Sprint(time.Now().Unix())
	nonce := rc.c.Query("nonce")
	token := rc.msgHandler.mpoptions.Token
	msgSignature := common.GenerateSignature(token, timestamp, nonce, encryptedMsg)

	var response string

	contentType := rc.c.GetHeader("Content-Type")
	if strings.Contains(contentType, "text/xml") {
		response = fmt.Sprintf(`<xml>
        <Encrypt><![CDATA[%s]]></Encrypt>
        <MsgSignature><![CDATA[%s]]></MsgSignature>
        <TimeStamp>%s</TimeStamp>
        <Nonce><![CDATA[%s]]></Nonce>
    </xml>`, encryptedMsg, msgSignature, timestamp, nonce)
	} else {
		// todo 这里需要转义
		response = fmt.Sprintf(`{"Encrypt":"%s","MsgSignature":"%s","TimeStamp":%s,"Nonce":"%s"}`, encryptedMsg, msgSignature, timestamp, nonce)
	}

	return []byte(response), nil
}

func (rc *ReplyCtrl) reply(data any) {
	var replyBytes []byte
	var err error

	contentType := rc.c.GetHeader("Content-Type")

	if strings.Contains(contentType, "text/xml") {
		replyBytes, err = xml.Marshal(data)
		if err != nil {
			rc.c.String(200, "success")
			return
		}
	} else {
		replyBytes, err = json.Marshal(data)
		if err != nil {
			rc.c.String(200, "success")
			return
		}
	}

	// log.Println("replyBytes---111", string(replyBytes))

	// 微信居然不支持回包是加密的，只能回包是明文的，是我搞错了么？
	// if rc.c.Query("encrypt_type") == "aes" {
	// 	replyBytes, err = rc.getEncryptBytes(replyBytes)
	// 	if err != nil {
	// 		rc.c.String(200, "success")
	// 		return
	// 	}
	// }

	// log.Println("replyBytes---222", string(replyBytes))

	if strings.Contains(contentType, "text/xml") {
		rc.c.Data(200, "application/xml", replyBytes)
	} else {
		rc.c.Data(200, "application/json", replyBytes)
	}
}

// 回复文本消息
func (rc *ReplyCtrl) ReplyText(content string) {
	reply := &ReplyMessageText{
		ToUserName:   rc.msg.GetFromUserName(),
		FromUserName: rc.msg.GetToUserName(),
		CreateTime:   time.Now().Unix(),
		MsgType:      "text",
		Content:      content,
	}
	rc.reply(reply)
}

// 回复图片消息
func (rc *ReplyCtrl) ReplyImage(mediaId string) {
	reply := &ReplyMessageImage{
		ToUserName:   rc.msg.GetFromUserName(),
		FromUserName: rc.msg.GetToUserName(),
		CreateTime:   time.Now().Unix(),
		MsgType:      "image",
		Image: SubStructMedia{
			MediaId: mediaId,
		},
	}
	rc.reply(reply)
}

// 回复语音消息
func (rc *ReplyCtrl) ReplyVoice(mediaId string) {
	reply := &ReplyMessageVoice{
		ToUserName:   rc.msg.GetFromUserName(),
		FromUserName: rc.msg.GetToUserName(),
		CreateTime:   time.Now().Unix(),
		MsgType:      "voice",
		Voice: SubStructMedia{
			MediaId: mediaId,
		},
	}
	rc.reply(reply)
}

// 回复视频消息
func (rc *ReplyCtrl) ReplyVideo(mediaId, title, description string) {
	reply := &ReplyMessageVideo{
		ToUserName:   rc.msg.GetFromUserName(),
		FromUserName: rc.msg.GetToUserName(),
		CreateTime:   time.Now().Unix(),
		MsgType:      "video",
		Video: SubStructMedia2{
			MediaId:     mediaId,
			Title:       title,
			Description: description,
		},
	}
	rc.reply(reply)
}

// 回复音乐消息
func (rc *ReplyCtrl) ReplyMusic(title, description, musicUrl, hqMusicUrl, thumbMediaId string) {
	reply := &ReplyMessageMusic{
		ToUserName:   rc.msg.GetFromUserName(),
		FromUserName: rc.msg.GetToUserName(),
		CreateTime:   time.Now().Unix(),
		MsgType:      "music",
		Music: SubStructMedia3{
			Title:        title,
			Description:  description,
			ThumbMediaId: thumbMediaId,
			MusicUrl:     musicUrl,
			HQMusicUrl:   hqMusicUrl,
		},
	}
	rc.reply(reply)
}

// 回复图文消息(外链)
func (rc *ReplyCtrl) ReplyNews(articles []*ReplyMessageArticle) {
	reply := &ReplyMessageNews{
		ToUserName:   rc.msg.GetFromUserName(),
		FromUserName: rc.msg.GetToUserName(),
		CreateTime:   time.Now().Unix(),
		MsgType:      "news",
		ArticleCount: len(articles),
		Articles: SubStructArticles{
			Item: articles,
		},
	}
	rc.reply(reply)
}

// 调用客服接口发送消息
func (rc *ReplyCtrl) send(data map[string]any) error {
	wxApiClient := rc.GetWxApiClient()
	if wxApiClient == nil {
		log.Println("ReplyCtrl.send, wxApiClient is nil, do nothing")
		return nil
	}
	ctx := context.TODO()
	err := wxApiClient.SendCustomMessage(ctx, rc.GetMsg().GetFromUserName(), data)
	if err != nil {
		log.Println("ReplyCtrl.send, wxApiClient.SendCustomMessage error", err)
		return err
	}
	return nil
}

// 发送文本
func (rc *ReplyCtrl) SendText(content string) error {
	data := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": content,
		},
	}
	return rc.send(data)
}

// 发送图片
func (rc *ReplyCtrl) SendImage(mediaId string) error {
	data := map[string]any{
		"msgtype": "image",
		"image": map[string]string{
			"media_id": mediaId,
		},
	}
	return rc.send(data)
}

// 发送语音
func (rc *ReplyCtrl) SendVoice(mediaId string) error {
	data := map[string]any{
		"msgtype": "voice",
		"voice": map[string]string{
			"media_id": mediaId,
		},
	}
	return rc.send(data)
}

// 发送视频
func (rc *ReplyCtrl) SendVideo(mediaId, thumbMediaId, title, description string) error {
	data := map[string]any{
		"msgtype": "video",
		"video": map[string]string{
			"media_id":       mediaId,
			"thumb_media_id": thumbMediaId,
			"title":          title,
			"description":    description,
		},
	}
	return rc.send(data)
}

// 发送音乐
func (rc *ReplyCtrl) SendMusic(title, description, musicUrl, hqMusicUrl, thumbMediaId string) error {
	data := map[string]any{
		"msgtype": "music",
		"music": map[string]string{
			"title":          title,
			"description":    description,
			"musicurl":       musicUrl,
			"hqmusicurl":     hqMusicUrl,
			"thumb_media_id": thumbMediaId,
		},
	}
	return rc.send(data)
}

// 发送图文(外链)
func (rc *ReplyCtrl) SendNews(articles []*SendMessageArticle) error {
	data := map[string]any{
		"msgtype": "news",
		"news": map[string]any{
			"articles": articles,
		},
	}
	return rc.send(data)
}

// 发送微信图文
// 这个方式，微信即将废弃
func (rc *ReplyCtrl) SendMpNews(mediaId string) error {
	data := map[string]any{
		"msgtype": "mpnews",
		"mpnews": map[string]any{
			"media_id": mediaId,
		},
	}
	return rc.send(data)
}

// 发送微信图文(文章)
func (rc *ReplyCtrl) SendMpNewsArticle(articleId string) error {
	data := map[string]any{
		"msgtype": "mpnewsarticle",
		"mpnewsarticle": map[string]any{
			"article_id": articleId,
		},
	}
	return rc.send(data)
}

// 发送卡券
func (rc *ReplyCtrl) SendWxCard(cardId string) error {
	data := map[string]any{
		"msgtype": "wxcard",
		"wxcard": map[string]string{
			"card_id": cardId,
		},
	}
	return rc.send(data)
}

// 发送小程序卡片
func (rc *ReplyCtrl) SendMiniProgramPage(title, appId, pagePath, thumbMediaId string) error {
	data := map[string]any{
		"msgtype": "miniprogrampage",
		"miniprogrampage": map[string]string{
			"title":          title,
			"appid":          appId,
			"pagepath":       pagePath,
			"thumb_media_id": thumbMediaId,
		},
	}
	return rc.send(data)
}

type MsgHandler struct {
	mpoptions   *mpoptions.MpOptions
	wxApiClient *wxapi.WxApi
	handler     MsgHandlerFunc
}

func defaultMsgHandlerFunc(rc *ReplyCtrl, msg Message) {
	rc.c.String(200, "success")
}

func NewMsgHandler(mpoptions *mpoptions.MpOptions, wxApiClient *wxapi.WxApi) *MsgHandler {
	return &MsgHandler{
		mpoptions:   mpoptions,
		wxApiClient: wxApiClient,
		handler:     defaultMsgHandlerFunc,
	}
}

func (m *MsgHandler) GetMpOptions() *mpoptions.MpOptions {
	return m.mpoptions
}

func (m *MsgHandler) SetMpOptions(mpoptions *mpoptions.MpOptions) {
	m.mpoptions = mpoptions
}

func (m *MsgHandler) GetWxApiClient() *wxapi.WxApi {
	return m.wxApiClient
}

func (m *MsgHandler) SetHandler(handler MsgHandlerFunc) {
	m.handler = handler
}

func (m *MsgHandler) returnFail(c *gin.Context, code int, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"code":    code,
		"message": msg,
	})
}

type EncryptMessage struct {
	XMLName    xml.Name `xml:"xml"`
	ToUserName string   `xml:"ToUserName" json:"ToUserName"`
	Encrypt    string   `xml:"Encrypt" json:"Encrypt"`
}

func (m *MsgHandler) getBody(c *gin.Context) ([]byte, error) {
	body, err := c.GetRawData()
	if err != nil {
		return nil, err
	}
	// log.Println("body", string(body))

	encrypt_type := c.Query("encrypt_type")

	if encrypt_type == "aes" {
		var encryptMsg EncryptMessage
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "text/xml") {
			err := xml.Unmarshal(body, &encryptMsg)
			if err != nil {
				log.Println("xml.Unmarshal error", err)
				return nil, err
			}
		} else {
			err := json.Unmarshal(body, &encryptMsg)
			if err != nil {
				log.Println("json.Unmarshal error", err)
				return nil, err
			}
		}

		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")
		msg_signature := c.Query("msg_signature")
		encrypt := encryptMsg.Encrypt

		params := []string{m.mpoptions.Token, timestamp, nonce, encrypt}
		sort.Strings(params)

		str := strings.Join(params, "")

		hash := sha1.New()
		hash.Write([]byte(str))
		hashedStr := hex.EncodeToString(hash.Sum(nil))

		if hashedStr != msg_signature {
			return nil, fmt.Errorf("msg_signature error")
		}

		retBody, err := common.AesDecryptWechat(m.mpoptions.AesKey, encrypt)

		if err != nil {
			return nil, err
		}

		body = retBody
	}

	return body, nil
}

func (m *MsgHandler) parseMessage(c *gin.Context) (Message, error) {
	var msg Message

	tmpMsg := struct {
		XMLName xml.Name `xml:"xml"`
		MsgType string   `xml:"MsgType" json:"MsgType"`
	}{}
	switch c.Request.Method {
	case "POST":
		body, err := m.getBody(c)
		if err != nil {
			log.Println("get body error", err)
			return nil, err
		}

		// bodyStr := string(body)
		// fmt.Println("bodyStr", bodyStr)

		contentType := c.GetHeader("Content-Type")

		if strings.Contains(contentType, "text/xml") {
			err := xml.Unmarshal(body, &tmpMsg)
			if err != nil {
				return nil, err
			}
		} else {
			err := json.Unmarshal(body, &tmpMsg)
			if err != nil {
				return nil, err
			}
		}

		switch tmpMsg.MsgType {
		case "text":
			msg = &MessageText{}
		case "image":
			msg = &MessageImage{}
		case "voice":
			msg = &MessageVoice{}
		case "video":
			msg = &MessageVideo{}
		case "shortvideo":
			msg = &MessageShortVideo{}
		case "location":
			msg = &MessageLocation{}
		case "link":
			msg = &MessageLink{}
		case "event":
			msg = &MessageEvent{}
		default:
			return nil, nil
		}

		if strings.Contains(contentType, "text/xml") {
			err = xml.Unmarshal(body, msg)
		} else {
			err = json.Unmarshal(body, msg)
		}
		if err != nil {
			return nil, err
		}
	default:
		return nil, nil
	}
	return msg, nil
}

func (m *MsgHandler) validateSignature(c *gin.Context) bool {
	signature := c.Query("signature")
	timestamp := c.Query("timestamp")
	nonce := c.Query("nonce")
	// echostr := c.Query("echostr")

	// 1. 将 token、timestamp、nonce 按字典序排序
	params := []string{m.mpoptions.Token, timestamp, nonce}
	sort.Strings(params)

	// 2. 将排序后的三个参数拼接成一个字符串
	str := strings.Join(params, "")

	// 3. 使用 SHA1 加密
	hash := sha1.New()
	hash.Write([]byte(str))
	hashedStr := hex.EncodeToString(hash.Sum(nil))

	// 4. 对比加密后的字符串与 signature
	return hashedStr == signature
}

func (m *MsgHandler) Serve(c *gin.Context) {
	var msg Message
	var err error

	if !m.validateSignature(c) {
		m.returnFail(c, 1, "signature error")
		return
	}

	switch c.Request.Method {
	case "POST":
		msg, err = m.parseMessage(c)
		if err != nil {
			m.returnFail(c, 1, err.Error())
			return
		}
	default:
		c.String(200, c.Query("echostr"))
		return
	}

	rc := &ReplyCtrl{
		c:          c,
		msgHandler: m,
		msg:        msg,
	}
	m.handler(rc, msg)
}
