//go:build !windows

package daemon

import (
	"path/filepath"
	"strings"
)

func isHiddenFile(path string) (bool, error) {
	return strings.HasPrefix(filepath.Base(path), "."), nil
}
