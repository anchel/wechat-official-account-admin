package types

import "time"

type ContextKey string
type ContextValue string

type SessionAppidInfo struct {
	AppType string `json:"app_type" bson:"app_type"` // 订阅号，公众号，小程序
	Name    string `json:"name" bson:"name"`

	AppID          string `json:"appid" bson:"appid"`
	AppSecret      string `json:"appsecret" bson:"appsecret"`
	Token          string `json:"token" bson:"token"`
	EncodingAESKey string `json:"encoding_aes_key" bson:"encoding_aes_key"`
}

type GinRequestLogInfo struct {
	AppID     string    `json:"appid" bson:"appid"`
	Time      time.Time `json:"time" bson:"time"`
	Status    int       `json:"status" bson:"status"`
	Latency   float64   `json:"latency" bson:"latency"`
	Ip        string    `json:"ip" bson:"ip"`
	Method    string    `json:"method" bson:"method"`
	Path      string    `json:"path" bson:"path"`
	Query     string    `json:"query" bson:"query"`
	Body      string    `json:"body" bson:"body"`
	UserAgent string    `json:"user-agent" bson:"user-agent"`
}
