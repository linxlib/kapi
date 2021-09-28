package internal

import (
	"os/exec"
	"runtime"
)

// Open calls the OS default program for uri
func OpenBrowser(uri string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	cmd := exec.Command("cmd", "/c", "start", uri)
	return cmd.Start()
}
