package mpoptions

type MpOptions struct {
	AppId     string
	AppSecret string
	Token     string
	AesKey    string
}

func NewMpOptions(appid, appsecret, token, aeskey string) *MpOptions {
	return &MpOptions{
		AppId:     appid,
		AppSecret: appsecret,
		Token:     token,
		AesKey:    aeskey,
	}
}
