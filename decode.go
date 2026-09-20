package morsekit

import (
	"errors"
	"fmt"
	"strings"
)

// ErrUnknownProsign is the sentinel wrapped by the errors DecodeMorse returns
// when a token in the code has no match in the Morse tables and IgnoreUnknown
// is false.
var ErrUnknownProsign = errors.New("morsekit: unknown code token")

// UnknownTokenError identifies the token that could not be decoded.
type UnknownTokenError struct {
	Token string
}

func (e *UnknownTokenError) Error() string {
	return fmt.Sprintf("%s %q", ErrUnknownProsign, e.Token)
}

func (e *UnknownTokenError) Unwrap() error { return ErrUnknownProsign }

// DecodeOpts controls DecodeMorse behaviour.
//
// The zero value gives: Upper = true (uppercase output), Separator = " ",
// IgnoreUnknown = false. Default-uppercase keeps EncodeText -> DecodeMorse
// round-trips exact.
type DecodeOpts struct {
	// Upper selects output case. nil means true (uppercase); point it at
	// false for lowercase output.
	Upper *bool
	// IgnoreUnknown replaces tokens missing from the tables with "?" so the
	// rest of the message still decodes. Otherwise DecodeMorse returns an
	// *UnknownTokenError.
	IgnoreUnknown bool
	// Separator joins decoded words. "" means " ".
	Separator string
}

func (o DecodeOpts) upper() bool {
	if o.Upper == nil {
		return true
	}
	return *o.Upper
}

func (o DecodeOpts) separator() string {
	if o.Separator == "" {
		return " "
	}
	return o.Separator
}

// DecodeMorse converts a Morse code string back to plain text.
//
// Tokens are split on whitespace (\s+). A lone "/" is a word gap; a run of
// dots and dashes is looked up in the code tables (so "-..-." is the letter
// slash, never a gap). Every other token - or any run that is not in the
// tables - is unknown. Consecutive spaces, tabs and newlines are tolerated.
func DecodeMorse(code string, opts DecodeOpts) (string, error) {
	sep := opts.separator()
	var words []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			words = append(words, cur.String())
			cur.Reset()
		}
	}
	for _, tok := range strings.Fields(code) {
		if tok == "/" {
			flush()
			continue
		}
		if !validSymbols(tok) {
			if opts.IgnoreUnknown {
				cur.WriteByte('?')
				continue
			}
			return "", &UnknownTokenError{Token: tok}
		}
		text, ok := decodeTable[tok]
		if !ok {
			if opts.IgnoreUnknown {
				cur.WriteByte('?')
				continue
			}
			return "", &UnknownTokenError{Token: tok}
		}
		if opts.upper() {
			text = strings.ToUpper(text)
		} else {
			text = strings.ToLower(text)
		}
		cur.WriteString(text)
	}
	flush()
	return strings.Join(words, sep), nil
}

// Roundtrip encodes text to Morse and decodes it back with default options,
// returning the text upper-case normalised (EncodeText folds letters to upper
// case).
func Roundtrip(text string) (string, error) {
	code, err := EncodeText(text)
	if err != nil {
		return "", err
	}
	return DecodeMorse(code, DecodeOpts{})
}

// validSymbols reports whether tok is a non-empty run of dots and dashes.
func validSymbols(tok string) bool {
	if tok == "" {
		return false
	}
	for i := 0; i < len(tok); i++ {
		if tok[i] != '.' && tok[i] != '-' {
			return false
		}
	}
	return true
}