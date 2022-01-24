package toml

import "time"

type Document struct {
	root *Node
}

func newDocument() Document {
	var output Document
	output.root = newNodePointer()
	output.root.kind = kindRoot
	return output
}

func (d Document) Section(path string) (*Node, bool) {
	return d.root.GetSection(path)
}

func (d Document) Value(path string) (Value, bool) {
	return d.root.GetValue(path)
}

func (d Document) ToString() string {
	return d.root.String()
}

func (d Document) Array(name string, defaultValue ...[]Value) []Value {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return make([]Value, 0)
		}
	}
	return v.AsArray()
}

func (d Document) String(name string, defaultValue ...string) string {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return ""
		}
	}
	return v.AsString()
}
func (d Document) Strings(name string, defaultValues ...string) []string {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValues) >= 1 {
			return defaultValues
		} else {
			return []string{}
		}
	}
	return v.AsStrings()
}

func (d Document) Int(name string, defaultValue ...int) int {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
	return v.AsInt()
}

func (d Document) Int8(name string, defaultValue ...int8) int8 {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
	return v.AsInt8()
}

func (d Document) Int16(name string, defaultValue ...int16) int16 {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
	return v.AsInt16()
}

func (d Document) Int32(name string, defaultValue ...int32) int32 {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
	return v.AsInt32()
}

func (d Document) Int64(name string, defaultValue ...int64) int64 {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0
		}
	}
	return v.AsInt64()
}

func (d Document) Float(name string, defaultValue ...float64) float64 {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0.0
		}
	}
	return v.AsFloat()
}

func (d Document) Float32(name string, defaultValue ...float32) float32 {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0.0
		}
	}
	return v.AsFloat32()
}

func (d Document) Float64(name string, defaultValue ...float64) float64 {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return 0.0
		}
	}
	return v.AsFloat64()
}

func (d Document) Bool(name string, defaultValue ...bool) bool {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return false
		}
	}
	return v.AsBool()
}

func (d Document) Date(name string, defaultValue ...time.Time) time.Time {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return time.Now()
		}
	}
	return v.AsDate()
}

func (d Document) Duration(name string, defaultValue ...time.Duration) time.Duration {
	v, ok := d.Value(name)
	if !ok {
		if len(defaultValue) >= 1 {
			return defaultValue[0]
		} else {
			return time.Second
		}
	}
	return v.AsDuration()
}
