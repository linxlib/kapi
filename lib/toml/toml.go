// A Go parser library for the [Toml format](https://github.com/mojombo/toml).

package toml

import (
	"io/ioutil"
)

type Kind int

var _globalParser = Parser{}

func ParseFile(tomlFilePath string) Document {
	content, err := ioutil.ReadFile(tomlFilePath)
	if err != nil {
		content = []byte("")
	}
	return _globalParser.Parse(string(content))
}
