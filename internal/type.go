package internal

import "strings"

var internalType = []string{"string", "bool", "int", "uint", "byte", "rune",
	"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "uintptr",
	"float32", "float64", "map", "Time", "error"}

// IsInternalType 是否是内部类型
func IsInternalType(t string) bool {
	for _, v := range internalType {
		if strings.EqualFold(t, v) {
			return true
		}
	}
	return false
}
