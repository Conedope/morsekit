package morsekit

import (
	"errors"
	"testing"
)

func boolPtr(b bool) *bool { return &b }

func TestDecodeHelloWorld(t *testing.T) {
	code := ".... . .-.. .-.. --- / .-- --- .-. .-.. -.."
	got, err := DecodeMorse(code, DecodeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "HELLO WORLD"; got != want {
		t.Errorf("decode = %q, want %q", got, want)
	}
}

func TestDecodeUnknownToken(t *testing.T) {
	unknowns := []string{"..-.-", ".-.-.-.", "x", "./-"}
	for _, tok := range unknowns {
		_, err := DecodeMorse(tok, DecodeOpts{})
		if err == nil {
			t.Errorf("DecodeMorse(%q): expected error", tok)
			continue
		}
		if !errors.Is(err, ErrUnknownProsign) {
			t.Errorf("DecodeMorse(%q) error %v does not wrap ErrUnknownProsign", tok, err)
		}
		var ut *UnknownTokenError
		if !errors.As(err, &ut) || ut.Token != tok {
			t.Errorf("DecodeMorse(%q): expected UnknownTokenError carrying %q", tok, tok)
		}
	}
}

func TestDecodeIgnoreUnknown(t *testing.T) {
	for _, code := range []string{".... ..-.- .", ".... x ."} {
		got, err := DecodeMorse(code, DecodeOpts{IgnoreUnknown: true})
		if err != nil {
			t.Fatalf("DecodeMorse(%q): %v", code, err)
		}
		if want := "H?E"; got != want {
			t.Errorf("DecodeMorse(%q) = %q, want %q", code, got, want)
		}
	}
}

func TestDecodeEmpty(t *testing.T) {
	got, err := DecodeMorse("", DecodeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("DecodeMorse(\"\") = %q, want \"\"", got)
	}
}

func TestDecodeConsecutiveSpaces(t *testing.T) {
	got, err := DecodeMorse("....    .\t /  .--", DecodeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "HE W"; got != want {
		t.Errorf("consecutive spaces = %q, want %q", got, want)
	}
}

func TestDecodeLower(t *testing.T) {
	got, err := DecodeMorse(".... . .-.. .-.. --- / .-- --- .-. .-.. -..",
		DecodeOpts{Upper: boolPtr(false)})
	if err != nil {
		t.Fatal(err)
	}
	if want := "hello world"; got != want {
		t.Errorf("lower decode = %q, want %q", got, want)
	}
	// nil means uppercase (documented default)
	got, err = DecodeMorse(".... .", DecodeOpts{})
	if err != nil || got != "HE" {
		t.Errorf("default decode = %q, %v; want \"HE\"", got, err)
	}
}

func TestDecodeSeparator(t *testing.T) {
	got, err := DecodeMorse(".... . / .-- .", DecodeOpts{Separator: " | "})
	if err != nil {
		t.Fatal(err)
	}
	if want := "HE | WE"; got != want {
		t.Errorf("separator decode = %q, want %q", got, want)
	}
}

func TestRoundtripTable(t *testing.T) {
	for _, ch := range TableChars() {
		got, err := Roundtrip(string(ch))
		if err != nil {
			t.Fatalf("Roundtrip(%q): %v", ch, err)
		}
		if got != string(ch) {
			t.Errorf("Roundtrip(%q) = %q", ch, got)
		}
	}
}

func TestRoundtripHelloWorld(t *testing.T) {
	got, err := Roundtrip("Hello World")
	if err != nil {
		t.Fatal(err)
	}
	if want := "HELLO WORLD"; got != want {
		t.Errorf("Roundtrip = %q, want %q", got, want)
	}
}

func TestRoundtripProsign(t *testing.T) {
	for tok := range prosignCodes {
		got, err := Roundtrip(tok)
		if err != nil {
			t.Fatalf("Roundtrip(%s): %v", tok, err)
		}
		if tok == "<KN>" {
			if want := "("; got != want {
				t.Errorf("KN shares -.--. with '('; expected %q, got %q", want, got)
			}
			continue
		}
		if got != tok {
			t.Errorf("Roundtrip(%s) = %q", tok, got)
		}
	}
}

func TestDecodeKNPriority(t *testing.T) {
	got, err := DecodeMorse("-.--.", DecodeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "("; got != want {
		t.Errorf("-.--. decodes to %q, want %q (punctuation wins over <KN>)", got, want)
	}
}

func TestDecodeTableSize(t *testing.T) {
	want := len(letterCodes) + len(digitCodes) + len(punctCodes) + len(prosignCodes) - 1
	if len(decodeTable) != want {
		t.Errorf("decodeTable has %d entries; want %d (one KN/( collision merged)", len(decodeTable), want)
	}
}

func TestDecodeWordGapOnlyDropsEmpty(t *testing.T) {
	got, err := DecodeMorse(".... / / .-- .", DecodeOpts{})
	if err != nil {
		t.Fatal(err)
	}
	if want := "H WE"; got != want {
		t.Errorf("double slash = %q, want %q", got, want)
	}
}