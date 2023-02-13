//go:build !windows

package daemon

import "os"

func sameFile(fi1, fi2 os.FileInfo) bool {
	return os.SameFile(fi1, fi2)
}
