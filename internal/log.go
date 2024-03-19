package internal

import (
	"time"
)

func Spend(name string) func() {
	t := time.Now()
	return func() {
		Infof("[%s] consume: %s", name, time.Since(t).String())
	}
}
