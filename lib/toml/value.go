package toml

import (
	"strconv"
	"strings"
	"time"
)

type Value struct {
	raw        string
	kind       Kind
	asBool     bool
	asInt      int64
	asFloat    float64
	asString   string
	asArray    []Value
	asDate     time.Time
	asDuration time.Duration
}

func (v Value) AsArray() []Value {
	return v.asArray
}

func (v Value) AsString() string {
	return v.asString
}

func (v Value) AsInt() int {
	return int(v.asInt)
}

func (v Value) AsInt8() int8 {
	return int8(v.asInt)
}

func (v Value) AsInt16() int16 {
	return int16(v.asInt)
}

func (v Value) AsInt32() int32 {
	return int32(v.asInt)
}

func (v Value) AsInt64() int64 {
	return v.asInt
}

func (v Value) AsFloat() float64 {
	return v.asFloat
}

func (v Value) AsFloat32() float32 {
	return float32(v.asFloat)
}

func (v Value) AsFloat64() float64 {
	return v.asFloat
}

func (v Value) AsBool() bool {
	return v.asBool
}

func (v Value) AsDate() time.Time {
	return v.asDate
}
func (v Value) AsDuration() time.Duration {
	return v.asDuration
}

func (v Value) String() string {
	if v.kind == kindString {
		s := v.asString
		s = strings.Replace(s, "\n", "\\n", -1)
		s = strings.Replace(s, "\x00", "\\0", -1)
		s = strings.Replace(s, "\t", "\\t", -1)
		s = strings.Replace(s, "\r", "\\r", -1)
		s = strings.Replace(s, "\"", "\\\"", -1)
		s = strings.Replace(s, "\\", "\\\\", -1)
		return "\"" + s + "\""
	}
	if v.kind == kindInt {
		return strconv.FormatInt(v.asInt, 10)
	}
	if v.kind == kindFloat {
		return strconv.FormatFloat(v.asFloat, 'f', -1, 64)
	}
	if v.kind == kindBool {
		if v.asBool {
			return "true"
		} else {
			return "false"
		}
	}
	if v.kind == kindDate {
		return v.asDate.Format(time.RFC3339)
	}
	if v.kind == kindArray {
		array := v.asArray
		output := ""
		for i := 0; i < len(array); i++ {
			if output != "" {
				output += ", "
			}
			output += array[i].String()
		}
		return "[" + output + "]"
	}
	return "undefined"
}
