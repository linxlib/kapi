package utils

import (
	"regexp"
	"sync"
)

var regexMu = sync.RWMutex{}
var regexMap = make(map[string]*regexp.Regexp)

// Replace do replace all matched <pattern> in bytes <src> with bytes <replace>.
func Replace(pattern string, replace, src []byte) ([]byte, error) {
	if r, err := getRegexp(pattern); err == nil {
		return r.ReplaceAll(src, replace), nil
	} else {
		return nil, err
	}
}

// ReplaceString replace all matched <pattern> in string <src> with string <replace>.
func ReplaceString(pattern, replace, src string) (string, error) {
	r, e := Replace(pattern, []byte(replace), []byte(src))
	return string(r), e
}

// getRegexp returns *regexp.Regexp object with given <pattern>.
// It uses cache to enhance the performance for compiling regular expression pattern,
// which means, it will return the same *regexp.Regexp object with the same regular
// expression pattern.
//
// It is concurrent-safe for multiple goroutines.
func getRegexp(pattern string) (regex *regexp.Regexp, err error) {
	// Retrieve the regular expression object using reading lock.
	regexMu.RLock()
	regex = regexMap[pattern]
	regexMu.RUnlock()
	if regex != nil {
		return
	}
	// If it does not exist in the cache,
	// it compiles the pattern and creates one.
	regex, err = regexp.Compile(pattern)
	if err != nil {
		return
	}
	// Cache the result object using writing lock.
	regexMu.Lock()
	regexMap[pattern] = regex
	regexMu.Unlock()
	return
}
