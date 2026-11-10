package frostui

import "testing"

func TestCStringCopiesTheText(t *testing.T) {
	b := cString("abc")
	if b == nil || *b != 'a' {
		t.Fatal("cString does not copy the text")
	}
}
