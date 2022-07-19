package rundaemon

var fatalSignals = []os.Signal{
	os.Interrupt,
	os.Kill,
}

var buildOutputName = "main_reload.exe"

func terminateGracefully(process *os.Process) error {
	return errors.New("terminateGracefully not implemented on windows")
}

func gracefulTerminationPossible() bool {
	return false
}
