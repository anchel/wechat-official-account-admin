package commonservice

import (
	"log"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
)

func GetPublicIP() (string, error) {
	// restyClient := utils.GetRestyClient()
	restyClient := resty.New()
	// 设置超时时间
	restyClient.SetTimeout(4 * time.Second)
	// 设置代理
	wxProxy := os.Getenv("WA_PROXY")
	if wxProxy != "" {
		restyClient.SetProxy(wxProxy)
	}

	req := restyClient.NewRequest()
	resp, err := req.Get("https://ifconfig.me/ip")
	if err != nil {
		log.Println("GetPublicIP", "err", err)
		return "", nil
	}

	return string(resp.Body()), nil
}
