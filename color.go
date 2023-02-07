package kapi

import (
	"fmt"
	"github.com/gookit/color"
	"time"
)

var (
	white        = color.HiWhite.Render
	red          = color.HiRed.Render
	green        = color.HiGreen.Render
	yellow       = color.HiYellow.Render
	magenta      = color.HiMagenta.Render
	lightmagenta = color.LightMagenta.Render

	kapiFormat = "[%s] %s "
	tinfo      = "I"
	twarn      = "W"
	tdebug     = "D"
	terr       = "E"
	tfatal     = "F"
	tok        = "√"
	tfail      = "x"
	tnone      = ""
)

func format(tag string) string {
	if tag == "" {
		return time.Now().Format("01-02 15:04:05")
	}
	return fmt.Sprintf(kapiFormat, tag, time.Now().Format("01-02 15:04:05"))
}

func Infof(fmt string, args ...any) {
	color.Printf(format(tinfo)+white(fmt)+"\n", args...)
}

func Whitef(fmt string, args ...any) {
	color.Printf(format(tnone)+white(fmt)+"\n", args...)
}
func Errorf(fmt string, args ...any) {
	color.Printf(format(terr)+red(fmt)+"\n", args...)
}
func Redf(fmt string, args ...any) {
	color.Printf(format(tnone)+red(fmt)+"\n", args...)
}
func Debugf(fmt string, args ...any) {
	color.Printf(format(tdebug)+green(fmt)+"\n", args...)
}
func Greenf(fmt string, args ...any) {
	color.Printf(format(tnone)+green(fmt)+"\n", args...)
}
func Warnf(fmt string, args ...any) {
	color.Printf(format(twarn)+yellow(fmt)+"\n", args...)
}
func Fatalf(fmt string, args ...any) {
	color.Printf(format(twarn)+lightmagenta(fmt)+"\n", args...)
}
func OKf(fmt string, args ...any) {
	color.Printf(format(tok)+green(fmt)+"\n", args...)
}
func Failf(fmt string, args ...any) {
	color.Printf(format(tfail)+magenta(fmt)+"\n", args...)
}
func Yellowf(fmt string, args ...any) {
	color.Printf(format(tnone)+yellow(fmt)+"\n", args...)
}
