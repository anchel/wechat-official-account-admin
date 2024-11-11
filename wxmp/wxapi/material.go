package wxapi

import (
	"context"
	"encoding/json"
	"log"

	"github.com/anchel/wechat-official-account-admin/lib/util"
	"github.com/go-resty/resty/v2"
)

type MaterialListItem struct {
	MediaId    string `json:"media_id"`
	Name       string `json:"name"`
	UpdateTime int64  `json:"update_time"`
	URL        string `json:"url"`
	Content    struct {
		NewsItem []*MessageNewsItem `json:"news_item"`
	} `json:"content,omitempty"`
}

type MaterialList struct {
	TotalCount int32               `json:"total_count"`
	ItemCount  int32               `json:"item_count"`
	Item       []*MaterialListItem `json:"item"`
}

// 获取永久素材列表
func (wxapi *WxApi) GetMaterialList(ctx context.Context, media_type string, offset int, count int) (*MaterialList, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"type":   media_type,
			"offset": offset,
			"count":  count,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/material/batchget_material")
	})
	if err != nil {
		return nil, err
	}

	retobj := &MaterialList{}
	err = json.Unmarshal(body, retobj)
	if err != nil {
		log.Println("GetMaterialList json.Unmarshal", err.Error())
		return nil, err
	}

	return retobj, nil
}

/**
 * 下载永久素材
 *
 */
func (wxapi *WxApi) DownloadMaterial(ctx context.Context, media_type, media_id string) ([]byte, map[string]string, error) {
	var err error
	body, res, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"media_id": media_id,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/material/get_material")
	})
	if err != nil {
		return body, nil, err
	}

	contentDisposition := res.Header().Get("Content-Disposition")
	fileExtension := getContentDispositionExt(contentDisposition)
	retMap := map[string]string{
		"extension": fileExtension,
	}

	if media_type == "video" {
		retobj := struct {
			Title       string `json:"title"`
			Description string `json:"description"`
			DownloadURL string `json:"down_url"`
		}{}
		err := json.Unmarshal(body, &retobj)
		if err != nil {
			return nil, retMap, err
		}
		retMap["title"] = retobj.Title
		retMap["description"] = retobj.Description
		retMap["download_url"] = retobj.DownloadURL
		if retMap["extension"] == "" {
			retMap["extension"] = util.GetExtensionFromUrl(retobj.DownloadURL)
		}

		body, err = util.GetUrlResultBody(retobj.DownloadURL)
		if err != nil {
			return nil, retMap, err
		}
	}

	log.Println("DownloadMaterial:", contentDisposition, fileExtension)

	// 注意有些返回的是二进制流，有些是json，需要外层处理
	return body, retMap, nil
}

type UploadMaterialResponse struct {
	MediaId      string `json:"media_id"`
	Url          string `json:"url"`
	CreatedAt    int64  `json:"created_at"`     // 临时素材会返回这个字段
	ThumbMediaId string `json:"thumb_media_id"` // 临时素材上传，且是thumb时，会返回这个字段，不会返回media_id
}

/**
 * 上传永久素材
 * description: 仅在视频类型时有效, "{"title":VIDEO_TITLE, "introduction":INTRODUCTION}"
 */
func (wxapi *WxApi) UploadMaterial(ctx context.Context, media_type string, filePath string, description map[string]string) (*UploadMaterialResponse, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		req.SetQueryParam("type", media_type)
		req.SetFile("media", filePath)
		if media_type == "video" {
			str, err := json.Marshal(description)
			if err != nil {
				log.Println("UploadMaterial json.Marshal", err.Error()) // 这里不结束，继续执行
			}
			body := map[string]string{
				"description": string(str),
			}
			req.SetFormData(body)
		}
		return req.Post("/cgi-bin/material/add_material")
	})
	if err != nil {
		return nil, err
	}

	log.Println("UploadMaterial:", string(body))

	retobj := &UploadMaterialResponse{}
	err = json.Unmarshal(body, retobj)
	if err != nil {
		log.Println("UploadMaterial json.Unmarshal", err.Error())
		return nil, err
	}

	return retobj, nil
}

// 删除永久素材
func (wxapi *WxApi) DeleteMaterial(ctx context.Context, media_id string) error {
	_, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		body := map[string]interface{}{
			"media_id": media_id,
		}
		req.SetBody(body)
		return req.Post("/cgi-bin/material/del_material")
	})
	return err
}

// 下载临时素材
func (wxapi *WxApi) DownloadTempMaterial(ctx context.Context, media_id string) ([]byte, map[string]string, error) {
	body, res, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		req.SetQueryParam("media_id", media_id)
		return req.Get("/cgi-bin/media/get")
	})
	if err != nil {
		return nil, nil, err
	}

	contentDisposition := res.Header().Get("Content-Disposition")
	fileExtension := getContentDispositionExt(contentDisposition)
	retMap := map[string]string{
		"extension": fileExtension,
	}

	log.Println("DownloadTempMaterial:", contentDisposition, fileExtension)

	// 尝试判断是不是video类型
	retobj := struct {
		VideoURL string `json:"video_url"`
	}{}
	err = json.Unmarshal(body, &retobj)
	if err != nil { // 如果发生错误，说明不是视频，是其他类型，则直接返回body
		return body, retMap, nil
	} else {
		body, err = util.GetUrlResultBody(retobj.VideoURL)
		if err != nil {
			return body, retMap, err
		}
	}

	return body, retMap, nil
}

// 上传临时素材
func (wxapi *WxApi) UploadTempMaterial(ctx context.Context, media_type string, filePath string) (*UploadMaterialResponse, error) {
	body, _, err := wxapi.CommonRequest(true, func(req *resty.Request) (*resty.Response, error) {
		req.SetQueryParam("type", media_type)
		req.SetFile("media", filePath)
		return req.Post("/cgi-bin/media/upload")
	})
	if err != nil {
		return nil, err
	}

	log.Println("UploadTempMaterial:", string(body))

	retobj := UploadMaterialResponse{}
	err = json.Unmarshal(body, &retobj)
	if err != nil {
		log.Println("UploadTempMaterial json.Unmarshal", err.Error())
		return nil, err
	}

	// 微信也太坑了，临时素材上传，且是thumb时，会返回thumb_media_id，不会返回media_id
	if media_type == "thumb" {
		retobj.MediaId = retobj.ThumbMediaId
	}

	return &retobj, nil
}
