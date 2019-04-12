package loremipsum

import (
	"math/rand"
	"strings"
)

var (
	// The canonical Lorem ipsum.
	phrase = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
`
	// Convert the phrase (string) into a rune-slice (characters)
	chars = []rune(phrase)
)

const (
	// Limit the byte slice to a maximum potential length.
	maximumRandomCharacters = 1e6
)

// "A random number of characters of Lorem Ipsum" Characters are kind of tricky
// in Go. We have to assume the test writer was referring to a UTF-8 character,
// e.g. potentially multi-byte, although there are none in the phrase..
func RandomCharacters() []byte {
	// Make a string builder to randomly write Lorem Ipsum's runes into.
	var builder strings.Builder

	// Iterate up to a random length, lower than the maximum.
	for c := 0; c < rand.Intn(maximumRandomCharacters); c++ {
		builder.WriteRune(chars[rand.Intn(len(chars))])
	}

	// Turn the string builder into byte slice, since we need one of those for writing later.
	return []byte(builder.String())
}
