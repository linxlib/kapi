package tools

import (
	"encoding/json"
)

// GetJSONStr obj to json string
func GetJSONStr(obj interface{}, isFormat bool) string {
	var b []byte
	if isFormat {
		b, _ = json.MarshalIndent(obj, "", "     ")
	} else {
		b, _ = json.Marshal(obj)
	}
	return string(b)
}
