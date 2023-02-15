package utils

import (
	"github.com/linxlib/logs"
	"os"
	"os/exec"
)

func RunGoTidy() {
	err := RunCommand("go", "mod", "tidy")
	if err != nil {
		logs.Error(err)
	}
}

func RunCommand(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	e := c.Start()
	if e != nil {
		logs.Error(e)
		return e
	}
	err := c.Wait()
	if err != nil {
		logs.Error(err)
		return err
	}
	return nil
}
