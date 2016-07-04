package net

import (
	"testing"
	"fmt"
)

func TestEscapeString(t *testing.T) {
	if escapeString("hello") != "hello" {
		print("Something went very wrong with escaping the word \"hello\"\n")
		t.Fail()
	}
	if escaped := escapeString("/"); escaped != "\034" {
		fmt.Printf("Did not escape path separator properly. Got %x instead of %x", escaped, "\034");
		t.Fail()
	}
}
