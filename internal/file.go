package internal

import (
	"io/ioutil"
	"os"
	"path"
)

// FileIsExist 检查目录是否存在
func FileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// BuildDir 创建目录
func BuildDir(absDir string) error {
	return os.MkdirAll(path.Dir(absDir), os.ModePerm) //生成多级目录
}

// WriteFile 写入文件
func WriteFile(fname string, src []byte, isClear bool) bool {
	_ = BuildDir(fname)
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !isClear {
		flag = os.O_CREATE | os.O_RDWR | os.O_APPEND
	}
	f, err := os.OpenFile(fname, flag, 0666)
	if err != nil {
		return false
	}
	_, _ = f.Write(src)
	_ = f.Close()
	return true
}

// ReadFile 读取文件
func ReadFile(fname string) []byte {
	src, _ := ioutil.ReadFile(fname)
	return src
}
