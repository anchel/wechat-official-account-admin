package wxapi

import (
	"regexp"
	"strings"
)

type MessageNewsItem struct {
	Title            string `json:"title"`
	ThumbMediaId     string `json:"thumb_media_id"`
	Author           string `json:"author"`
	Digest           string `json:"digest"`
	ShowCoverPic     int    `json:"show_cover_pic"`
	Content          string `json:"content"`
	Url              string `json:"url"`
	ContentSourceUrl string `json:"content_source_url"`
}

func getContentDispositionExt(contentDisposition string) string {
	re := regexp.MustCompile(`filename="(.+)"`)
	matches := re.FindStringSubmatch(contentDisposition)

	if len(matches) > 1 {
		filename := matches[1]
		// 获取后缀名
		parts := strings.Split(filename, ".")
		var fileExtension string
		if len(parts) > 1 {
			fileExtension = parts[len(parts)-1] // 获取最后一个部分作为后缀名
		} else {
			return ""
		}

		return "." + fileExtension
	}
	return ""
}
