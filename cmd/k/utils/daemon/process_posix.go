//go:build !windows

package daemon

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
