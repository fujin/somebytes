package loremipsum

import (
	"strings"
	"testing"
)

// This is not a useful test.
func TestChars(t *testing.T) {
	if len(Chars) != 447 {
		t.Fatal("Lorem ipsum rune slice did not have expected length.")
	}
}

// Also not great.
func TestPhrase(t *testing.T) {
	if !strings.Contains(Phrase, "Lorem ipsum") {
		t.Fatal("Phrase did not contain expected string")
	}
}
