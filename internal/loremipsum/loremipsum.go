package loremipsum

import (
	"math/rand"
	"strings"

	"github.com/pkg/errors"
)

const (
	// The canonical Lorem ipsum.
	phrase = `
Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.
`
	// Limit the byte slice to a maximum potential length.
	maximumRandomCharacters = 1e6
)

var chars = []rune(phrase)

// RandomCharacters returns "A random number of characters of Lorem Ipsum" as a
// slice of bytes. Characters are kind of tricky in Go. We have to assume the
// test writer was referring to a UTF-8 character, e.g. potentially multi-byte,
// although there are none in the phrase..
func RandomCharacters() ([]byte, error) {
	// Make a string builder to randomly write Lorem Ipsum's runes into.
	var builder strings.Builder

	// Iterate up to a random length, lower than the maximum.
	for c := 0; c < rand.Intn(maximumRandomCharacters); c++ {
		_, err := builder.WriteRune(chars[rand.Intn(len(chars))])
		if err != nil {
			return nil, errors.Wrap(err, "could not write rune into string builder")
		}
	}

	// Turn the string builder into byte slice, since we need one of those for writing later.
	return []byte(builder.String()), nil
}
