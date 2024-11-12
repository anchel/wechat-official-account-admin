package commonservice

import "github.com/anchel/wechat-official-account-admin/lib/utils"

func GetPublicIP() (string, error) {
	restyClient := utils.GetRestyClient()

	req := restyClient.NewRequest()
	resp, err := req.Get("https://ipinfo.io/ip")
	if err != nil {
		return "", err
	}

	return string(resp.Body()), nil
}
