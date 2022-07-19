//go:build darwin || dragonfly || freebsd || linux || nacl || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux nacl netbsd openbsd solaris

package rundaemon

import (
	"os"
	"syscall"
)

var buildOutputName = "main_reload"

var fatalSignals = []os.Signal{
	syscall.SIGINT,
	syscall.SIGTERM,
	syscall.SIGQUIT,
}

func terminateGracefully(process *os.Process) error {
	return process.Signal(syscall.SIGTERM)
}

func gracefulTerminationPossible() bool {
	return true
}
