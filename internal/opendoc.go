package internal

import (
	"os/exec"
	"runtime"
)

// OpenBrowser Open calls the OS default program for uri
func OpenBrowser(uri string) error {
	if runtime.GOOS != "windows" {
		return nil
	}
	var args []string
	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}
	cmd := exec.Command(args[0], append(args[1:], uri)...)
	return cmd.Start()
}
