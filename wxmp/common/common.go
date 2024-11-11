package common

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/go-resty/resty/v2"
)

func GetAccessToken(appid, appsecret string) (string, int64, error) {
	client := resty.New()

	wxProxy := os.Getenv("WA_PROXY")
	if wxProxy != "" {
		log.Println("WA_PROXY:", wxProxy)
		client.SetProxy(wxProxy)
	}

	client.SetBaseURL("https://api.weixin.qq.com/cgi-bin")

	// ret := map[string]any{}
	ret := struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
		Errcode     int64  `json:"errcode"`
		Errmsg      string `json:"errmsg"`
	}{}
	_, err := client.R().
		SetQueryParams(map[string]string{
			"grant_type": "client_credential",
			"appid":      appid,
			"secret":     appsecret,
		}).
		SetResult(&ret).
		Get("/token")

	if err != nil {
		return "", 0, err
	}

	// log.Println(resp.StatusCode(), resp.Status(), resp.String())

	if ret.Errcode != 0 {
		return "", 0, fmt.Errorf("%d-%s", ret.Errcode, ret.Errmsg)
	}

	// log.Println("get access token:", ret["access_token"], ret["expires_in"])

	return ret.AccessToken, ret.ExpiresIn, nil
}

/**
 * 这个不是单纯的解密，还涉及到解密后的数据处理，即提取指定区间的字节数据
 */
func AesDecryptWechat(aesKey, encryptedMessage string) ([]byte, error) {
	// Base64 解码 AES 密钥
	aesKeyDecoded, err := base64.StdEncoding.DecodeString(aesKey + "=") // EncodingAESKey 加密需要补一个 '='
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 key: %v", err)
	}

	// Base64 解码密文
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedMessage)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64 encrypted message: %v", err)
	}

	// 使用密钥的前 16 字节作为 IV
	iv := aesKeyDecoded[:aes.BlockSize]

	// 创建 AES 解密器
	block, err := aes.NewCipher(aesKeyDecoded)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %v", err)
	}

	// 检查密文长度
	if len(ciphertext) < aes.BlockSize || len(ciphertext)%aes.BlockSize != 0 {
		return nil, errors.New("ciphertext length is invalid")
	}

	mode := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(ciphertext))
	mode.CryptBlocks(decrypted, ciphertext)

	// 去除 PKCS#7 填充
	decrypted = pkcs7Unpad(decrypted)

	// 去除前 16 字节长度的网络字节序，以及 AppID 后多余的字符
	contentLength := int(decrypted[16])<<24 | int(decrypted[17])<<16 | int(decrypted[18])<<8 | int(decrypted[19])
	// message := string(decrypted[20 : 20+contentLength])

	return decrypted[20 : 20+contentLength], nil
}

// PKCS#7 填充去除
func pkcs7Unpad(p []byte) []byte {
	padding := int(p[len(p)-1])
	return p[:len(p)-padding]
}

func AesEncrypt(plainText []byte, aesKey string) (string, error) {
	aesKeyDecoded, err := base64.StdEncoding.DecodeString(aesKey + "=") // EncodingAESKey 加密需要补一个 '='
	if err != nil {
		return "", fmt.Errorf("error decoding base64 key: %v", err)
	}

	block, err := aes.NewCipher(aesKeyDecoded)
	if err != nil {
		return "", err
	}

	// 使用AES加密
	iv := aesKeyDecoded[:aes.BlockSize]
	plainText = pkcs7Padding(plainText, aes.BlockSize)
	cipherText := make([]byte, len(plainText))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(cipherText, plainText)

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

// PKCS#7 填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func GenerateSignature(params ...string) string {
	sort.Strings(params)
	h := sha1.New()
	io.WriteString(h, strings.Join(params, ""))
	return fmt.Sprintf("%x", h.Sum(nil))
}
