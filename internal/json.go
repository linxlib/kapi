package internal

import (
	"encoding/json"
)

// MarshalToJson obj to json string
func MarshalToJson(obj interface{}, isFormat bool) string {
	var b []byte
	if isFormat {
		b, _ = json.MarshalIndent(obj, "", "     ")
	} else {
		b, _ = json.Marshal(obj)
	}
	return string(b)
}
