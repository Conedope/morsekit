package morsekit

import (
	"errors"
	"strings"
	"testing"
)

func TestEncodeSOS(t *testing.T) {
	got, err := EncodeText("SOS")
	if err != nil {
		t.Fatal(err)
	}
	if want := "... --- ..."; got != want {
		t.Errorf("SOS = %q, want %q", got, want)
	}
}

func TestEncodeHelloWorld(t *testing.T) {
	want := ".... . .-.. .-.. --- / .-- --- .-. .-.. -.."
	for _, in := range []string{"HELLO WORLD", "hello world", "Hello   World"} {
		got, err := EncodeText(in)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("EncodeText(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEncodePunctuation(t *testing.T) {
	cases := map[rune]string{
		'.': ".-.-.-", ',': "--..--", '?': "..--..", '\'': ".----.",
		'!': "-.-.--", '/': "-..-.", '@': ".--.-.", '(': "-.--.", ')': "-.--.-",
		'-': "-....-",
	}
	for ch, want := range cases {
		got, err := EncodeText(string(ch))
		if err != nil {
			t.Fatalf("EncodeText(%q): %v", ch, err)
		}
		if got != want {
			t.Errorf("EncodeText(%q) = %q, want %q", ch, got, want)
		}
	}
}

func TestEncodePunctuationSentence(t *testing.T) {
	got, err := EncodeText("CQ?")
	if err != nil {
		t.Fatal(err)
	}
	if want := "-.-. --.- ..--.."; got != want {
		t.Errorf("CQ? = %q, want %q", got, want)
	}
}

func TestEncodeProsign(t *testing.T) {
	for tok, want := range prosignCodes {
		got, err := EncodeText(tok)
		if err != nil {
			t.Fatalf("EncodeText(%s): %v", tok, err)
		}
		if got != want {
			t.Errorf("EncodeText(%s) = %q, want %q", tok, got, want)
		}
	}
}

func TestEncodeProsignInlineAndCase(t *testing.T) {
	got, err := EncodeText("HELLO<AR>CQ")
	if err != nil {
		t.Fatal(err)
	}
	if want := ".... . .-.. .-.. --- .-.-. -.-. --.-"; got != want {
		t.Errorf("HELLO<AR>CQ = %q, want %q", got, want)
	}
	got, err = EncodeText("<ar>")
	if err != nil {
		t.Fatal(err)
	}
	if got != ".-.-." {
		t.Errorf("<ar> = %q, want .-.-.", got)
	}
}

func TestEncodeUnsupported(t *testing.T) {
	for _, in := range []string{"#", "É", "\u00e9", "_", "\u263a", "A#B", "<XX>", "A<"} {
		_, err := EncodeText(in)
		if err == nil {
			t.Errorf("EncodeText(%q): expected an error", in)
			continue
		}
		if !errors.Is(err, ErrUnsupportedChar) {
			t.Errorf("EncodeText(%q) error %v does not wrap ErrUnsupportedChar", in, err)
		}
	}
}

func TestEncodeUnsupportedNamesChar(t *testing.T) {
	_, err := EncodeText("AB#CD")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "#") {
		t.Errorf("error %q should name the offending character", err)
	}
}

func TestEncodeEmpty(t *testing.T) {
	got, err := EncodeText("")
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("EncodeText(\"\") = %q, want \"\"", got)
	}
}

func TestEncodeCaseFolding(t *testing.T) {
	a, err := EncodeText("Hello WOrld")
	if err != nil {
		t.Fatal(err)
	}
	b, err := EncodeText("HELLO WORLD")
	if err != nil {
		t.Fatal(err)
	}
	if a != b {
		t.Errorf("case folding mismatch: %q vs %q", a, b)
	}
}

func TestEncodeDigits(t *testing.T) {
	got, err := EncodeText("12345")
	if err != nil {
		t.Fatal(err)
	}
	if want := ".---- ..--- ...-- ....- ....."; got != want {
		t.Errorf("12345 = %q, want %q", got, want)
	}
}