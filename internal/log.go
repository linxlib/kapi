package internal

import (
	"time"
)

func Spend(name string) func() {
	t := time.Now()
	return func() {
		Infof("[%s] time: %s", name, time.Since(t).String())
	}
}
