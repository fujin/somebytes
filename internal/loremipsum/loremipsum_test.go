package loremipsum

import (
	"strings"
	"testing"
)

// This is not a useful test.
func TestChars(t *testing.T) {
	if len(chars) != 447 {
		t.Fatal("Lorem ipsum rune slice did not have expected length.")
	}
}

// Also not great.
func TestPhrase(t *testing.T) {
	if !strings.Contains(phrase, "Lorem ipsum") {
		t.Fatal("Phrase did not contain expected string")
	}
}

func TestRandomCharacters(t *testing.T) {
	if string(RandomCharacters()) == phrase {
		t.Fatal("Random Characters was the phrase! That's not random at all.")
	}

	if len(RandomCharacters()) > maximumRandomCharacters {
		t.Fatal("Random characters over the maximum limit")
	}
}
