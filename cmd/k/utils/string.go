package utils

import "strings"

var (
	// DefaultTrimChars are the characters which are stripped by Trim* functions in default.
	DefaultTrimChars = string([]byte{
		'\t', // Tab.
		'\v', // Vertical tab.
		'\n', // New line (line feed).
		'\r', // Carriage return.
		'\f', // New page.
		' ',  // Ordinary space.
		0x00, // NUL-byte.
		0x85, // Delete.
		0xA0, // Non-breaking space.
	})
)

const (
	// NotFoundIndex is the position index for string not found in searching functions.
	NotFoundIndex = -1
)

func SplitAndTrim(str, delimiter string, characterMask ...string) []string {
	array := make([]string, 0)
	for _, v := range strings.Split(str, delimiter) {
		v = Trim(v, characterMask...)
		if v != "" {
			array = append(array, v)
		}
	}
	return array
}

func Trim(str string, characterMask ...string) string {
	return trim(str, characterMask...)
}

// Trim strips whitespace (or other characters) from the beginning and end of a string.
// The optional parameter <characterMask> specifies the additional stripped characters.
func trim(str string, characterMask ...string) string {
	trimChars := DefaultTrimChars
	if len(characterMask) > 0 {
		trimChars += characterMask[0]
	}
	return strings.Trim(str, trimChars)
}

// SearchArray searches string `s` in string slice `a` case-sensitively,
// returns its index in `a`.
// If `s` is not found in `a`, it returns -1.
func SearchArray(a []string, s string) int {
	for i, v := range a {
		if s == v {
			return i
		}
	}
	return NotFoundIndex
}

// InArray checks whether string `s` in slice `a`.
func InArray(a []string, s string) bool {
	return SearchArray(a, s) != NotFoundIndex
}

// Contains reports whether <substr> is within <str>, case-sensitively.
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}
