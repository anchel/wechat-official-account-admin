package utils

import (
	"log"
	"os"

	"github.com/go-resty/resty/v2"
)

var restyClient *resty.Client

func GetRestyClient() *resty.Client {
	if restyClient == nil {
		restyClient = resty.New()
		wxProxy := os.Getenv("WA_PROXY")
		if wxProxy != "" {
			log.Println("GetHttpClient WA_PROXY:", wxProxy)
			restyClient.SetProxy(wxProxy)
		}
	}
	return restyClient
}
