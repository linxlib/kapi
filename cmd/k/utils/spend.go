package utils

import (
	"github.com/gookit/color"
	"time"
)

func TimeSpend(name string) func() {
	t := time.Now()
	return func() {
		color.Debug.Printf("%s spend %s\n", name, time.Since(t))
	}
}
