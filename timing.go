package morsekit

import (
	"errors"
	"math"
	"strings"
)

// Result keys for AnalyzePacing.
const (
	PacingSymbols = "symbols"
	PacingDots    = "dots"
	PacingDashes  = "dashes"
	PacingLetters = "letters"
	PacingWords   = "words"
)

// ErrBadWPM reports a non-positive (or non-finite) words-per-minute value.
var ErrBadWPM = errors.New("morsekit: wpm must be positive and finite")

// AnalyzePacing analyses a code string token by token and returns counts of
// symbols (dots + dashes), dots, dashes, letters (symbol runs) and words
// (non-empty groups of letters separated by "/"). Empty or whitespace-only
// input returns an all-zero map. A token that is neither "/" nor a run of
// dots and dashes returns an *UnknownTokenError.
func AnalyzePacing(code string) (map[string]int, error) {
	res := map[string]int{
		PacingSymbols: 0, PacingDots: 0, PacingDashes: 0,
		PacingLetters: 0, PacingWords: 0,
	}
	inWord := false
	for _, tok := range strings.Fields(code) {
		switch {
		case tok == "/":
			inWord = false
		case validSymbols(tok):
			res[PacingLetters]++
			if !inWord {
				res[PacingWords]++
				inWord = true
			}
			for i := 0; i < len(tok); i++ {
				res[PacingSymbols]++
				if tok[i] == '.' {
					res[PacingDots]++
				} else {
					res[PacingDashes]++
				}
			}
		default:
			return nil, &UnknownTokenError{Token: tok}
		}
	}
	return res, nil
}

// ElementDurations returns the millisecond length of each Morse timing
// element for a given speed, per the PARIS/ITU standard.
//
// Timing formula: the word "PARIS" plus its trailing word gap totals 50
// units. Element lengths in units: dot = 1, dash = 3, intra-character
// gap = 1, letter gap = 3, word gap = 7. One unit lasts 1200/wpm ms, so at
// 20 wpm a dot is 60 ms, a dash 180 ms, the intra-character gap 60 ms, the
// letter gap 180 ms and the word gap 420 ms.
//
// units is an optional base-unit override in milliseconds: pass 0 to derive
// the unit from wpm. When units <= 0, wpm must be positive and finite or
// ErrBadWPM is returned.
func ElementDurations(units int, wpm float64) (dotMS, dashMS, intraGapMS, letterGapMS, wordGapMS float64, err error) {
	var unit float64
	if units > 0 {
		unit = float64(units)
	} else {
		if wpm <= 0 || math.IsNaN(wpm) || math.IsInf(wpm, 0) {
			return 0, 0, 0, 0, 0, ErrBadWPM
		}
		unit = 1200.0 / wpm
	}
	dotMS = unit
	dashMS = 3 * unit
	intraGapMS = unit
	letterGapMS = 3 * unit
	wordGapMS = 7 * unit
	return dotMS, dashMS, intraGapMS, letterGapMS, wordGapMS, nil
}