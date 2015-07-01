package utils

import (
	"runtime/debug"
	"testing"
)

func check(err error, t *testing.T) {
	if err != nil {
		debug.PrintStack()
		t.Fatal(err)
	}
}
