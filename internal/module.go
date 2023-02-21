package internal

import (
	"bufio"
	"errors"
	"os"
	"runtime"
	"strings"
)

// GetModuleInfo 获取项目[module name] [根目录绝对地址]
func GetModuleInfo(n int) (string, string, bool) {
	index := n
	// 本包被引用时需要向上推2级查找main.go
	for {
		_, filename, _, ok := runtime.Caller(index)
		if ok {
			if strings.HasSuffix(filename, "runtime/asm_amd64.s") {
				index -= 2
				break
			}
			if strings.HasSuffix(filename, "runtime/asm_arm64.s") {
				index -= 2
				break
			}
			index++
		} else {
			panic(errors.New("package parsing failed:can not find main file"))
		}
	}

	_, filename, _, _ := runtime.Caller(index)
	filename = strings.Replace(filename, "\\", "/", -1) // change windows path delimiter '\' to unix path delimiter '/'
	for {
		n := strings.LastIndex(filename, "/")
		if n > 0 {
			filename = filename[0:n]
			if FileIsExist(filename + "/go.mod") {
				return getMod(filename + "/go.mod"), filename, true
			}
		} else {
			break
		}
	}

	// never reach
	return "", "", false
}
func getMod(fileName string) string {
	file, err := os.Open(fileName)
	if err != nil {
		return ""
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
