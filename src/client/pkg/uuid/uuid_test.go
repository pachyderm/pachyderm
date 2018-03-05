package uuid

import (
	"testing"
)

func TestIsUUIDWithoutDashes(t *testing.T) {
	if len(NewWithoutDashes()) != 32 {
		t.Error()
	}

	if !IsUUIDWithoutDashes("09abcd098faa4fd98643023485739adb") {
		t.Error()
	}

	// 13 character is 4
	if IsUUIDWithoutDashes("09abcd098faaefd98643023485739adb") {
		t.Fail()
	}

	// Length 32
	if IsUUIDWithoutDashes("09abcd098faaefd98643023485739adbabc") {
		t.Fail()
	}

	// Hexadecimal
	if IsUUIDWithoutDashes("09abcd098faa4fd98643023485739xyz") {
		t.Fail()
	}

	// Generated
	if !IsUUIDWithoutDashes(NewWithoutDashes()) {
		t.Fail()
	}
}
