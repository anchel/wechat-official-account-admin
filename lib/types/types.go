package types

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
