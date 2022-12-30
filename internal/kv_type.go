package internal

var kvType = map[string]string{ // array, boolean, integer, number, object, string
	"int":     "integer",
	"uint":    "integer",
	"byte":    "integer",
	"rune":    "integer",
	"int8":    "integer",
	"int16":   "integer",
	"int32":   "integer",
	"int64":   "integer",
	"uint8":   "integer",
	"uint16":  "integer",
	"uint32":  "integer",
	"uint64":  "integer",
	"uintptr": "integer",
	"float32": "integer",
	"float64": "integer",
	"bool":    "boolean",
	"map":     "object",
	"string":  "string",
	"Time":    "string",
}

var kvFormat = map[string]string{}

// GetKvType 获取类型转换
func GetKvType(k string, isArray, isType bool) string {
	if isArray {
		return "array"
	}

	if isType {
		if kt, ok := kvType[k]; ok {
			return kt
		}
		return "object"
	}
	if kf, ok := kvFormat[k]; ok {
		return kf
	}
	return k
}
