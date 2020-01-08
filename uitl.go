package app

import (
	"net"
	"os"
	"path/filepath"
	"strings"
)

// 获取绝对路径
func GetAbs(filename string) string {
	if !filepath.IsAbs(filename) {
		filename, _ = filepath.Abs(filename)
	}
	return filename
}

// 获取当前路径
func GetCurrPath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

// 目录存在否
func IsDirExist(dir string) bool {
	fi, err := os.Stat(dir)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}

// 获取本机IP
func GetLocalAddr() (string, error) {
	var localIp string
	if conn, err := net.Dial("udp", "baidu.com:80"); err != nil {
		return "", err
	} else {
		localIp = strings.Split(conn.LocalAddr().String(), ":")[0]
		_ = conn.Close()
		return localIp, nil
	}
}
