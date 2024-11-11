package util

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func GetExePwd() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exePath), nil
}

func GetUniqueFileName(prefix string) string {
	objectID := primitive.NewObjectID()
	return prefix + objectID.Hex()
}

func GetFileNameByMediaId(prefix, suffix, media_id string) string {
	// 创建一个新的 SHA-1 哈希
	h := sha1.New()

	// 写入字符串的字节
	h.Write([]byte(media_id))

	// 计算哈希值
	sha1Hash := h.Sum(nil)

	// 将哈希值转换为十六进制字符串
	return prefix + hex.EncodeToString(sha1Hash) + suffix
}

func GetUploadFilePath(suffix string) (string, string, error) {
	wd, err := GetExePwd()
	if err != nil {
		return "", "", err
	}
	filename := GetUniqueFileName("upload-") + suffix
	dstFilePath := filepath.Join(wd, "files", "upload-image", filename)
	filePath := filepath.Join("/files", "upload-image", filename)
	return dstFilePath, filePath, nil
}

func GetWxDownloadMediaFilePath(prefix, suffix, media_id string) (string, string, error) {
	wd, err := GetExePwd()
	if err != nil {
		return "", "", err
	}
	filename := GetFileNameByMediaId(prefix, suffix, media_id)
	dstFilePath := filepath.Join(wd, "files", "wx-download-media", filename)
	filePath := filepath.Join("/files", "wx-download-media", filename)
	return dstFilePath, filePath, nil
}

func GetUrlPathByFilePath(filePath string) string {
	wd, err := GetExePwd()
	if err != nil {
		return ""
	}
	return strings.Replace(filePath, wd, "", 1)

}

func GetExtByMediaType(media_type string) string {
	switch media_type {
	case "image":
		return ".jpg"
	case "voice":
		return ".mp3"
	case "video":
		return ".mp4"
	case "thumb":
		return ".jpg"
	default:
		return ""
	}
}

func SaveFile(absPath string, data []byte) error {
	// 获取文件所在的目录
	dir := filepath.Dir(absPath)

	// 创建缺失的文件夹
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("无法创建目录: %v", err)
	}

	// 创建文件
	err = os.WriteFile(absPath, data, os.ModePerm)
	if err != nil {
		return fmt.Errorf("无法写入文件: %v", err)
	}

	return nil
}

func FileExistsAndAccessible(path string) (bool, error) {
	// 尝试打开文件
	file, err := os.Open(path)
	if err != nil {
		// 如果错误类型是文件不存在，返回 false 和 nil 错误
		if os.IsNotExist(err) {
			return false, nil
		}
		// 如果文件存在，但不能访问，返回 false 和具体错误
		return false, err
	}
	defer file.Close()
	// 如果没有错误，文件存在且可以访问
	return true, nil
}

func GetUrlResultBody(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func GetExtensionFromStr(input string) string {
	// 查找最后一个 '.' 的位置
	lastDot := strings.LastIndex(input, ".")

	// 如果没有找到 '.'，返回空字符串
	if lastDot == -1 || lastDot == len(input)-1 {
		return ""
	}

	// 提取扩展名
	return input[lastDot:]
}

func GetExtensionFromUrl(url string) string {
	// 获取最后的部分
	filename := filepath.Base(url)

	// 获取扩展名
	ext := filepath.Ext(filename)

	return ext
}
