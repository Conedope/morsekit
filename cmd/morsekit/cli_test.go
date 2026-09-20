package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var bin string

func TestMain(m *testing.M) {
	tmp, err := os.MkdirTemp("", "morsekit-cli")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	bin = filepath.Join(tmp, "morsekit")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building CLI: %v\n%s", err, out)
		os.RemoveAll(tmp)
		os.Exit(1)
	}
	code := m.Run()
	os.RemoveAll(tmp)
	os.Exit(code)
}

// runCLI runs the built binary and returns stdout, stderr and the exit code.
func runCLI(t *testing.T, stdin string, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	if stdin != "" {
		cmd.Stdin = strings.NewReader(stdin)
	}
	var so, se strings.Builder
	cmd.Stdout = &so
	cmd.Stderr = &se
	err := cmd.Run()
	code := 0
	if err != nil {
		var ee *exec.ExitError
		if !errors.As(err, &ee) {
			t.Fatalf("running %v: %v", args, err)
		}
		code = ee.ExitCode()
	}
	return so.String(), se.String(), code
}

func TestCLIEncodeDecodeRoundtrip(t *testing.T) {
	enc, _, code := runCLI(t, "", "encode", "Hello, World!")
	if code != 0 {
		t.Fatalf("encode exit %d", code)
	}
	wantEnc := ".... . .-.. .-.. --- --..-- / .-- --- .-. .-.. -.. -.-.--"
	if strings.TrimSpace(enc) != wantEnc {
		t.Errorf("encode = %q, want %q", strings.TrimSpace(enc), wantEnc)
	}

	dec, _, code := runCLI(t, "", "decode", strings.TrimSpace(enc))
	if code != 0 {
		t.Fatalf("decode exit %d", code)
	}
	if strings.TrimSpace(dec) != "HELLO, WORLD!" {
		t.Errorf("decode = %q, want HELLO, WORLD!", strings.TrimSpace(dec))
	}

	lower, _, code := runCLI(t, "", "decode", "--lower", strings.TrimSpace(enc))
	if code != 0 {
		t.Fatalf("decode --lower exit %d", code)
	}
	if strings.TrimSpace(lower) != "hello, world!" {
		t.Errorf("decode --lower = %q, want hello, world!", strings.TrimSpace(lower))
	}
}

func TestCLIHelloWorldSample(t *testing.T) {
	dec, _, code := runCLI(t, "", "decode", ".... . .-.. .-.. --- / .-- --- .-. .-.. -..")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if strings.TrimSpace(dec) != "HELLO WORLD" {
		t.Errorf("decode = %q, want HELLO WORLD", strings.TrimSpace(dec))
	}
}

func TestCLITiming(t *testing.T) {
	out, se, code := runCLI(t, "", "timing",
		".... . .-.. .-.. --- / .-- --- .-. .-.. -..", "--wpm", "20")
	if code != 0 {
		t.Fatalf("timing exit %d, stderr: %s", code, se)
	}
	for _, want := range []string{
		"symbols: 32", "dots: 19", "dashes: 13", "letters: 10", "words: 2",
		"wpm: 20", "60.0 ms", "180.0 ms", "420.0 ms", "1200/wpm",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("timing output missing %q:\n%s", want, out)
		}
	}
}

func TestCLIDecodeUnknownExit1(t *testing.T) {
	_, se, code := runCLI(t, "", "decode", ".... .-.-.-.")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(se, "unknown") {
		t.Errorf("stderr = %q, want mention of unknown token", se)
	}
}

func TestCLIDecodeIgnoreUnknown(t *testing.T) {
	out, _, code := runCLI(t, "", "decode", "--ignore-unknown", ".... .-.-.-.")
	if code != 0 {
		t.Errorf("exit = %d, want 0", code)
	}
	if strings.TrimSpace(out) != "H?" {
		t.Errorf("--ignore-unknown decode = %q, want H?", strings.TrimSpace(out))
	}
}

func TestCLIEncodeEmpty(t *testing.T) {
	out, _, code := runCLI(t, "", "encode", "")
	if code != 0 {
		t.Errorf("empty encode exit = %d, want 0", code)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("empty encode stdout = %q, want \"\"", out)
	}
}

func TestCLIEncodeEmptyStdin(t *testing.T) {
	out, _, code := runCLI(t, "", "encode")
	if code != 0 {
		t.Errorf("empty stdin encode exit = %d, want 0", code)
	}
	if strings.TrimSpace(out) != "" {
		t.Errorf("empty stdin encode stdout = %q, want \"\"", out)
	}
}

func TestCLIEncodeUnsupportedExit1(t *testing.T) {
	_, se, code := runCLI(t, "", "encode", "#")
	if code != 1 {
		t.Errorf("exit = %d, want 1", code)
	}
	if !strings.Contains(se, "unsupported") {
		t.Errorf("stderr = %q, want unsupported character error", se)
	}
}

func TestCLIVersion(t *testing.T) {
	out, _, code := runCLI(t, "", "--version")
	if code != 0 {
		t.Fatalf("--version exit %d", code)
	}
	if !strings.HasPrefix(strings.TrimSpace(out), "morsekit ") {
		t.Errorf("--version = %q, want morsekit <ver>", out)
	}
	sub, _, code := runCLI(t, "", "encode", "--version")
	if code != 0 || !strings.HasPrefix(strings.TrimSpace(sub), "morsekit ") {
		t.Errorf("encode --version = %q (exit %d)", sub, code)
	}
}

func TestCLIHelp(t *testing.T) {
	out, _, code := runCLI(t, "", "--help")
	if code != 0 {
		t.Fatalf("--help exit %d", code)
	}
	for _, want := range []string{"encode", "decode", "timing", "--wpm", "--ignore-unknown"} {
		if !strings.Contains(out, want) {
			t.Errorf("help missing %q", want)
		}
	}
}

func TestCLIUsageErrors(t *testing.T) {
	if _, _, code := runCLI(t, ""); code != 2 {
		t.Errorf("no args exit = %d, want 2", code)
	}
	if _, _, code := runCLI(t, "", "bogus"); code != 2 {
		t.Errorf("unknown command exit = %d, want 2", code)
	}
	if _, _, code := runCLI(t, "", "encode", "--bogus", "x"); code != 2 {
		t.Errorf("unknown flag exit = %d, want 2", code)
	}
	if _, _, code := runCLI(t, "", "timing", ".-", "--wpm"); code != 2 {
		t.Errorf("missing --wpm value exit = %d, want 2", code)
	}
	if _, _, code := runCLI(t, "", "timing", ".-", "--wpm", "abc"); code != 2 {
		t.Errorf("bad --wpm value exit = %d, want 2", code)
	}
}

func TestCLITimingBadWPM(t *testing.T) {
	if _, _, code := runCLI(t, "", "timing", ".-", "--wpm", "0"); code != 1 {
		t.Errorf("wpm 0 exit = %d, want 1", code)
	}
}

func TestCLIStdin(t *testing.T) {
	out, _, code := runCLI(t, "SOS", "encode")
	if code != 0 {
		t.Fatalf("stdin encode exit %d", code)
	}
	if strings.TrimSpace(out) != "... --- ..." {
		t.Errorf("stdin encode = %q, want ... --- ...", strings.TrimSpace(out))
	}
}