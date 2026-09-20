package morsekit

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ErrUnsupportedChar is the sentinel wrapped by the errors EncodeText returns
// for characters (or prosign tokens) that have no Morse code.
var ErrUnsupportedChar = errors.New("morsekit: unsupported character")

type unsupportedCharError struct {
	ch rune
}

func (e *unsupportedCharError) Error() string {
	return fmt.Sprintf("morsekit: unsupported character %q (U+%04X)", e.ch, e.ch)
}

func (e *unsupportedCharError) Unwrap() error { return ErrUnsupportedChar }

// EncodeText converts plain text to a Morse code string.
//
// Documented decisions:
//   - Text is treated as an ASCII stream read rune by rune. ASCII letters,
//     upper or lower case, are folded to upper case and mapped via the
//     letter table; non-ASCII characters have no Morse form and fail with an
//     error that names the character and wraps ErrUnsupportedChar.
//   - Whitespace runs are word boundaries. Letters inside a word are joined
//     by a single space; words are separated by " / ".
//   - Prosign tokens such as "<AR>" are recognised case-insensitively
//     (<AR>, <ar>, <aR> all work) and expanded to their symbol run. An
//     unknown or unterminated <...> token fails as an unsupported character.
//   - Any other character is an unsupported-character error.
//   - EncodeText("") returns "" with no error.
func EncodeText(text string) (string, error) {
	words := strings.Fields(text)
	codes := make([]string, 0, len(words))
	for _, w := range words {
		code, err := encodeWord(w)
		if err != nil {
			return "", err
		}
		codes = append(codes, code)
	}
	return strings.Join(codes, " / "), nil
}

func encodeWord(w string) (string, error) {
	syms := make([]string, 0, len(w))
	for i := 0; i < len(w); {
		if w[i] == '<' {
			j := strings.IndexByte(w[i+1:], '>')
			if j < 0 {
				return "", fmt.Errorf("%w: unterminated prosign token in %q", ErrUnsupportedChar, w)
			}
			tok := strings.ToUpper(w[i : i+j+2])
			code, ok := prosignCodes[tok]
			if !ok {
				return "", fmt.Errorf("%w: unknown prosign token %q", ErrUnsupportedChar, tok)
			}
			syms = append(syms, code)
			i += j + 2
			continue
		}
		r, size := utf8.DecodeRuneInString(w[i:])
		if code, ok := letterCodes[unicode.ToUpper(r)]; ok {
			syms = append(syms, code)
			i += size
			continue
		}
		if code, ok := digitCodes[r]; ok {
			syms = append(syms, code)
			i += size
			continue
		}
		if code, ok := punctCodes[r]; ok {
			syms = append(syms, code)
			i += size
			continue
		}
		return "", &unsupportedCharError{ch: r}
	}
	return strings.Join(syms, " "), nil
}