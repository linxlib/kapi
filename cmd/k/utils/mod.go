package utils

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/linxlib/logs"
	"os"
	"os/exec"
	"strings"
)

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

func GetGoVersion() (string, string, string) {
	buf := bytes.Buffer{}

	c := exec.Command("go", "env", "-json")
	c.Stdout = &buf
	c.Stderr = os.Stderr
	e := c.Start()
	if e != nil {
		logs.Error(e)
		return "", "", ""
	}
	err := c.Wait()
	if err != nil {
		logs.Error(err)
		return "", "", ""
	}
	m := make(map[string]any)
	err = json.Unmarshal(buf.Bytes(), &m)
	if err != nil {
		logs.Error(err)
		return "", "", ""
	}
	return m["GOVERSION"].(string), m["GOHOSTOS"].(string), m["GOHOSTARCH"].(string)
}
