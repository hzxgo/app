package app

import (
	"net"
	"os"
	"path/filepath"
	"strings"
)

func GetAbs(filename string) string {
	if !filepath.IsAbs(filename) {
		filename, _ = filepath.Abs(filename)
	}
	return filename
}

func GetCurrPath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func IsDirExist(dir string) bool {
	fi, err := os.Stat(dir)
	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}

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
