package internal

import "testing"

func TestGetMod(t *testing.T) {
	a:=GetMod("../go.mod")
	t.Log(a)
}
