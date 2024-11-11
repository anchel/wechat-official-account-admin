package msghandler

import "encoding/xml"

type MessageText struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	Content      string `xml:"Content" json:"Content"`
	MsgId        int64  `xml:"MsgId" json:"MsgId"`
	MsgDataId    string `xml:"MsgDataId" json:"MsgDataId"`
	Idx          string `xml:"Idx" json:"Idx"`
}

func (m *MessageText) GetMsgType() string {
	return "text"
}
func (m *MessageText) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageText) GetFromUserName() string {
	return m.FromUserName
}

type MessageImage struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	PicUrl       string `xml:"PicUrl" json:"PicUrl"`
	MediaId      string `xml:"MediaId" json:"MediaId"`
	MsgId        int64  `xml:"MsgId" json:"MsgId"`
	MsgDataId    string `xml:"MsgDataId" json:"MsgDataId"`
	Idx          string `xml:"Idx" json:"Idx"`
}

func (m *MessageImage) GetMsgType() string {
	return "image"
}
func (m *MessageImage) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageImage) GetFromUserName() string {
	return m.FromUserName
}

type MessageVoice struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	MediaId      string `xml:"MediaId" json:"MediaId"`
	Format       string `xml:"Format" json:"Format"`
	MsgId        int64  `xml:"MsgId" json:"MsgId"`
	MsgDataId    string `xml:"MsgDataId" json:"MsgDataId"`
	Idx          string `xml:"Idx" json:"Idx"`
	MediaId16K   string `xml:"MediaId16K" json:"MediaId16K"`
}

func (m *MessageVoice) GetMsgType() string {
	return "voice"
}
func (m *MessageVoice) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageVoice) GetFromUserName() string {
	return m.FromUserName
}

type MessageVideo struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	MediaId      string `xml:"MediaId" json:"MediaId"`
	ThumbMediaId string `xml:"ThumbMediaId" json:"ThumbMediaId"`
	MsgId        int64  `xml:"MsgId" json:"MsgId"`
	MsgDataId    string `xml:"MsgDataId" json:"MsgDataId"`
	Idx          string `xml:"Idx" json:"Idx"`
}

func (m *MessageVideo) GetMsgType() string {
	return "video"
}
func (m *MessageVideo) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageVideo) GetFromUserName() string {
	return m.FromUserName
}

type MessageShortVideo struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	MediaId      string `xml:"MediaId" json:"MediaId"`
	ThumbMediaId string `xml:"ThumbMediaId" json:"ThumbMediaId"`
	MsgId        int64  `xml:"MsgId" json:"MsgId"`
	MsgDataId    string `xml:"MsgDataId" json:"MsgDataId"`
	Idx          string `xml:"Idx" json:"Idx"`
}

func (m *MessageShortVideo) GetMsgType() string {
	return "shortvideo"
}
func (m *MessageShortVideo) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageShortVideo) GetFromUserName() string {
	return m.FromUserName
}

type MessageLocation struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	LocationX    string `xml:"Location_X" json:"Location_X"`
	LocationY    string `xml:"Location_Y" json:"Location_Y"`
	Scale        string `xml:"Scale" json:"Scale"`
	Label        string `xml:"Label" json:"Label"`
	MsgId        int64  `xml:"MsgId" json:"MsgId"`
	MsgDataId    string `xml:"MsgDataId" json:"MsgDataId"`
	Idx          string `xml:"Idx" json:"Idx"`
}

func (m *MessageLocation) GetMsgType() string {
	return "location"
}
func (m *MessageLocation) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageLocation) GetFromUserName() string {
	return m.FromUserName
}

type MessageLink struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	Title        string `xml:"Title" json:"Title"`
	Description  string `xml:"Description" json:"Description"`
	Url          string `xml:"Url" json:"Url"`
	MsgId        int64  `xml:"MsgId" json:"MsgId"`
	MsgDataId    string `xml:"MsgDataId" json:"MsgDataId"`
	Idx          string `xml:"Idx" json:"Idx"`
}

func (m *MessageLink) GetMsgType() string {
	return "link"
}
func (m *MessageLink) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageLink) GetFromUserName() string {
	return m.FromUserName
}

type MessageEvent struct {
	XMLName xml.Name `xml:"xml"`

	ToUserName   string `xml:"ToUserName" json:"ToUserName"`
	FromUserName string `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64  `xml:"CreateTime" json:"CreateTime"`
	MsgType      string `xml:"MsgType" json:"MsgType"`
	Event        string `xml:"Event" json:"Event"`
	EventKey     string `xml:"EventKey" json:"EventKey"`
	Ticket       string `xml:"Ticket" json:"Ticket"`
}

func (m *MessageEvent) GetMsgType() string {
	return "event"
}
func (m *MessageEvent) GetToUserName() string {
	return m.ToUserName
}
func (m *MessageEvent) GetFromUserName() string {
	return m.FromUserName
}

/**
 * 回复微信服务器的相关消息结构体
 */

type SubStructMedia struct {
	MediaId string `xml:"MediaId" json:"MediaId"`
}

type SubStructMedia2 struct {
	MediaId     string `xml:"MediaId" json:"MediaId"`
	Title       string `xml:"Title" json:"Title"`
	Description string `xml:"Description" json:"Description"`
}

type SubStructMedia3 struct {
	Title        string `xml:"Title" json:"Title"`
	Description  string `xml:"Description" json:"Description"`
	ThumbMediaId string `xml:"ThumbMediaId" json:"ThumbMediaId"`
	MusicUrl     string `xml:"MusicUrl" json:"MusicUrl"`
	HQMusicUrl   string `xml:"HQMusicUrl" json:"HQMusicUrl"`
}

type ReplyMessageText struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName" json:"ToUserName"`
	FromUserName string   `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime" json:"CreateTime"`
	MsgType      string   `xml:"MsgType" json:"MsgType"`
	Content      string   `xml:"Content" json:"Content"`
}

type ReplyMessageImage struct {
	XMLName      xml.Name       `xml:"xml"`
	ToUserName   string         `xml:"ToUserName" json:"ToUserName"`
	FromUserName string         `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64          `xml:"CreateTime" json:"CreateTime"`
	MsgType      string         `xml:"MsgType" json:"MsgType"`
	Image        SubStructMedia `xml:"Image" json:"Image"`
}

type ReplyMessageVoice struct {
	XMLName      xml.Name       `xml:"xml"`
	ToUserName   string         `xml:"ToUserName" json:"ToUserName"`
	FromUserName string         `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64          `xml:"CreateTime" json:"CreateTime"`
	MsgType      string         `xml:"MsgType" json:"MsgType"`
	Voice        SubStructMedia `xml:"Voice" json:"Voice"`
}

type ReplyMessageVideo struct {
	XMLName      xml.Name        `xml:"xml"`
	ToUserName   string          `xml:"ToUserName" json:"ToUserName"`
	FromUserName string          `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64           `xml:"CreateTime" json:"CreateTime"`
	MsgType      string          `xml:"MsgType" json:"MsgType"`
	Video        SubStructMedia2 `xml:"Video" json:"Video"`
}

type ReplyMessageMusic struct {
	XMLName      xml.Name        `xml:"xml"`
	ToUserName   string          `xml:"ToUserName" json:"ToUserName"`
	FromUserName string          `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64           `xml:"CreateTime" json:"CreateTime"`
	MsgType      string          `xml:"MsgType" json:"MsgType"`
	Music        SubStructMedia3 `xml:"Music" json:"Music"`
}

type ReplyMessageArticle struct {
	Title       string `xml:"Title" json:"Title"`
	Description string `xml:"Description" json:"Description"`
	PicUrl      string `xml:"PicUrl" json:"PicUrl"`
	Url         string `xml:"Url" json:"Url"`
}

type SubStructArticles struct {
	Item []*ReplyMessageArticle `xml:"item" json:"item"`
}

type ReplyMessageNews struct {
	XMLName      xml.Name          `xml:"xml"`
	ToUserName   string            `xml:"ToUserName" json:"ToUserName"`
	FromUserName string            `xml:"FromUserName" json:"FromUserName"`
	CreateTime   int64             `xml:"CreateTime" json:"CreateTime"`
	MsgType      string            `xml:"MsgType" json:"MsgType"`
	ArticleCount int               `xml:"ArticleCount" json:"ArticleCount"`
	Articles     SubStructArticles `xml:"Articles" json:"Articles"`
}

type SendMessageArticle struct {
	Title       string `xml:"title" json:"title"`
	Description string `xml:"description" json:"description"`
	PicUrl      string `xml:"picurl" json:"picurl"`
	Url         string `xml:"url" json:"url"`
}
