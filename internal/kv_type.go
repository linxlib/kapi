package internal

func GetFormat(t string, customFormat string) string {
	switch t {
	case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "byte", "rune", "uintptr":
		return "int32"
	case "int64", "uint64":
		return "int64"
	case "bool":
		return ""
	case "string":
		return "string"
	case "float32", "float64":
		return "float"
	case "map":
		return ""
	case "Time":
		return "date-time"
	case "Date":
		return "date"
	case "array", "slice":
		return "object"
	default:
		return ""
	}
}

func GetType(t string) string {
	switch t {
	case "int", "uint", "int8", "uint8", "int16", "uint16", "int32", "uint32", "byte", "rune", "uintptr":
		return "integer"
	case "int64", "uint64":
		return "integer"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	case "float32", "float64":
		return "number"
	case "map":
		return "object"
	case "Time":
		return "string"
	case "Date":
		return "string"
	case "array", "slice":
		return "array"
	default:
		return "object"
	}
}
