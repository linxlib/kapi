package internal

import (
	"bufio"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
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

// GetPathDirs 获取目录所有文件夹
func GetPathDirs(absDir string) (re []string) {
	if FileIsExist(absDir) {
		files, _ := ioutil.ReadDir(absDir)
		for _, f := range files {
			if f.IsDir() {
				re = append(re, f.Name())
			}
		}
	}
	return
}

// GetCurrentDirectory 获取exe所在目录
func GetCurrentDirectory() string {
	dir, _ := os.Executable()
	exPath := filepath.Dir(dir)
	// fmt.Println(exPath)

	return exPath
}

// WriteFile 写入文件
func WriteFile(fname string, src []byte, isClear bool) bool {
	BuildDir(fname)
	flag := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !isClear {
		flag = os.O_CREATE | os.O_RDWR | os.O_APPEND
	}
	f, err := os.OpenFile(fname, flag, 0666)
	if err != nil {
		return false
	}
	defer f.Close()
	f.Write(src)

	return true
}

// ReadFile 读取文件
func ReadFile(fname string) []byte {
	src, _ := ioutil.ReadFile(fname)
	return src
}

func GetMod(fileName string) string {
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		m := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(m, "module") {
			m = strings.TrimPrefix(m, "module")
			m = strings.TrimSpace(m)
			return m
		}
	}
	return ""
}
