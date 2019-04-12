package loremipsum

import (
	"testing"
)

func TestRandomCharacters(t *testing.T) {
	chars, err := RandomCharacters()

	if err != nil {
		t.Fatal(err)
	}

	if len(chars) == 0 {
		t.Fatal("did not generate any random characters")
	}

	if len(chars) > maximumRandomCharacters {
		t.Fatal("generated more random characters than should have")
	}
}
