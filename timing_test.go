package morsekit

import (
	"errors"
	"reflect"
	"testing"
)

func wantMap(t *testing.T, got map[string]int, want map[string]int) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("result has %d keys, want %d: %v", len(got), len(want), got)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("result = %v, want %v", got, want)
	}
}

func TestAnalyzePacingParis(t *testing.T) {
	got, err := AnalyzePacing(".--. .- .-. .. ...")
	if err != nil {
		t.Fatal(err)
	}
	wantMap(t, got, map[string]int{
		PacingSymbols: 14, PacingDots: 10, PacingDashes: 4,
		PacingLetters: 5, PacingWords: 1,
	})
}

func TestAnalyzePacingHelloWorld(t *testing.T) {
	got, err := AnalyzePacing(".... . .-.. .-.. --- / .-- --- .-. .-.. -..")
	if err != nil {
		t.Fatal(err)
	}
	wantMap(t, got, map[string]int{
		PacingSymbols: 32, PacingDots: 19, PacingDashes: 13,
		PacingLetters: 10, PacingWords: 2,
	})
}

func TestAnalyzePacingEmpty(t *testing.T) {
	got, err := AnalyzePacing("  \t ")
	if err != nil {
		t.Fatal(err)
	}
	wantMap(t, got, map[string]int{
		PacingSymbols: 0, PacingDots: 0, PacingDashes: 0,
		PacingLetters: 0, PacingWords: 0,
	})
}

func TestAnalyzePacingInvalid(t *testing.T) {
	for _, bad := range []string{".... x", "./-", ".-..a"} {
		if _, err := AnalyzePacing(bad); !errors.Is(err, ErrUnknownProsign) {
			t.Errorf("AnalyzePacing(%q) error = %v, want ErrUnknownProsign", bad, err)
		}
	}
}

func TestElementDurationsWPM20(t *testing.T) {
	dot, dash, intra, letter, word, err := ElementDurations(0, 20)
	if err != nil {
		t.Fatal(err)
	}
	if dot != 60 || dash != 180 || intra != 60 || letter != 180 || word != 420 {
		t.Errorf("wpm=20 => %g %g %g %g %g, want 60 180 60 180 420",
			dot, dash, intra, letter, word)
	}
}

func TestElementDurationsChangesWithWPM(t *testing.T) {
	dot, _, _, _, _, err := ElementDurations(0, 40)
	if err != nil {
		t.Fatal(err)
	}
	if dot != 30 {
		t.Errorf("wpm=40 dot = %g, want 30", dot)
	}
}

func TestElementDurationsBadWPM(t *testing.T) {
	for _, wpm := range []float64{0, -20, 0.001 * -1} {
		if _, _, _, _, _, err := ElementDurations(0, wpm); !errors.Is(err, ErrBadWPM) {
			t.Errorf("ElementDurations(0, %v) error = %v, want ErrBadWPM", wpm, err)
		}
	}
	if _, _, _, _, _, err := ElementDurations(0, -1); err == nil {
		t.Error("negative wpm should error")
	}
}

func TestElementDurationsUnitsOverride(t *testing.T) {
	dot, dash, _, letter, word, err := ElementDurations(100, -999)
	if err != nil {
		t.Fatal(err)
	}
	if dot != 100 || dash != 300 || letter != 300 || word != 700 {
		t.Errorf("units=100 => %g %g %g %g, want 100 300 300 700", dot, dash, letter, word)
	}
}