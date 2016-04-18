package net

import (
	"testing"
)

func TestEscapeString(t *testing.T) {
	if escapeString("hello") != "hello" {
		print("Something went very wrong with escaping the word \"hello\"\n")
		t.Fail()
	}
	if escaped := escapeString("/"); escaped != "\\/" {
		print("Did not escape path separator properly. Got " + escaped + " instead of \\/ \n")
		t.Fail()
	}
}
