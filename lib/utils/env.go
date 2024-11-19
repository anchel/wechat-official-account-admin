package utils

import (
	"os"
	"path/filepath"
)

// 检查可执行文件所在目录下，.env 文件是否存在
func CheckEnvFile() bool {
	exePath, err := os.Executable()
	if err != nil {
		return false
	}
	exeDir := filepath.Dir(exePath)
	envPath := filepath.Join(exeDir, ".env")

	f, err := os.Stat(envPath)
	if os.IsNotExist(err) {
		return false
	}
	if err == nil && f.IsDir() {
		return false
	}
	return err == nil
}
