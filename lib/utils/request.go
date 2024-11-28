package utils

import (
	"os"

	"github.com/anchel/wechat-official-account-admin/lib/logger"
	"github.com/go-resty/resty/v2"
)

var restyClient *resty.Client

func GetRestyClient() *resty.Client {
	if restyClient == nil {
		restyClient = resty.New()
		wxProxy := os.Getenv("WA_PROXY")
		if wxProxy != "" {
			logger.Info("GetHttpClient", "WA_PROXY", wxProxy)
			restyClient.SetProxy(wxProxy)
		}
	}
	return restyClient
}
