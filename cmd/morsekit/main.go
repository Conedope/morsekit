package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/Conedope/morsekit"
)

const version = "0.1.0"

const usageText = `morsekit %s - Morse code toolkit (ITU-R M.1677-1, offline)

Usage:
  morsekit encode [flags] [TEXT]     text -> Morse code
  morsekit decode [flags] [CODE]     Morse code -> text (uppercase by default)
  morsekit timing [flags] [CODE]     symbol counts + element durations (PARIS)
  morsekit --version | --help

Flags:
  -f, --file FILE       read input from FILE instead of arguments/stdin
      --lower           decode: output lowercase text
      --ignore-unknown  decode: replace unknown tokens with "?" and exit 0
      --wpm N           timing: words per minute (default 20; unit = 1200/wpm ms)
  -h, --help            show this help
      --version         show version
  -                     as the only argument: read from stdin

Exit status: 0 success, 1 data error, 2 usage error.

Input resolution: CODE/TEXT arguments (joined with spaces) are used when
given; otherwise --file FILE is read; otherwise stdin is read.
`

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintf(stderr, usageText, version)
		return 2
	}
	switch args[0] {
	case "encode", "decode", "timing":
		f, st := parseFlags(args[1:])
		switch st {
		case parseErr:
			fmt.Fprintf(stderr, "morsekit: invalid arguments for %q\n\n", args[0])
			fmt.Fprintf(stderr, usageText, version)
			return 2
		case parseHelp:
			fmt.Fprintf(stdout, usageText, version)
			return 0
		case parseVersion:
			fmt.Fprintf(stdout, "morsekit %s\n", version)
			return 0
		}
		switch args[0] {
		case "encode":
			return cmdEncode(f, stdin, stdout, stderr)
		case "decode":
			return cmdDecode(f, stdin, stdout, stderr)
		default:
			return cmdTiming(f, stdin, stdout, stderr)
		}
	case "--version", "version":
		fmt.Fprintf(stdout, "morsekit %s\n", version)
		return 0
	case "--help", "-h", "help":
		fmt.Fprintf(stdout, usageText, version)
		return 0
	default:
		fmt.Fprintf(stderr, "morsekit: unknown command %q\n\n", args[0])
		fmt.Fprintf(stderr, usageText, version)
		return 2
	}
}

type flagSet struct {
	file          string
	lower         bool
	ignoreUnknown bool
	wpm           float64
	wpmSet        bool
	positional    []string
}

type parseStatus int

const (
	parseOK parseStatus = iota
	parseHelp
	parseVersion
	parseErr
)

func parseFlags(args []string) (*flagSet, parseStatus) {
	f := &flagSet{}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--help" || a == "-h":
			return f, parseHelp
		case a == "--version":
			return f, parseVersion
		case a == "--lower":
			f.lower = true
		case a == "--ignore-unknown":
			f.ignoreUnknown = true
		case a == "--file" || a == "-f":
			i++
			if i >= len(args) {
				return f, parseErr
			}
			f.file = args[i]
		case strings.HasPrefix(a, "--file="):
			f.file = strings.TrimPrefix(a, "--file=")
			if f.file == "" {
				return f, parseErr
			}
		case a == "--wpm":
			i++
			if i >= len(args) {
				return f, parseErr
			}
			v, err := strconv.ParseFloat(args[i], 64)
			if err != nil {
				return f, parseErr
			}
			f.wpm, f.wpmSet = v, true
		case strings.HasPrefix(a, "--wpm="):
			v, err := strconv.ParseFloat(strings.TrimPrefix(a, "--wpm="), 64)
			if err != nil {
				return f, parseErr
			}
			f.wpm, f.wpmSet = v, true
		default:
			if strings.HasPrefix(a, "-") && a != "-" {
				return f, parseErr
			}
			f.positional = append(f.positional, a)
		}
	}
	return f, parseOK
}

func readInput(f *flagSet, stdin io.Reader) (string, error) {
	if len(f.positional) > 0 && !(len(f.positional) == 1 && f.positional[0] == "-") {
		return strings.Join(f.positional, " "), nil
	}
	if f.file != "" {
		data, err := os.ReadFile(f.file)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func cmdEncode(f *flagSet, stdin io.Reader, stdout, stderr io.Writer) int {
	text, err := readInput(f, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "morsekit: encode: %v\n", err)
		return 1
	}
	code, err := morsekit.EncodeText(text)
	if err != nil {
		fmt.Fprintf(stderr, "morsekit: encode: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, code)
	return 0
}

func cmdDecode(f *flagSet, stdin io.Reader, stdout, stderr io.Writer) int {
	code, err := readInput(f, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "morsekit: decode: %v\n", err)
		return 1
	}
	opts := morsekit.DecodeOpts{IgnoreUnknown: f.ignoreUnknown}
	if f.lower {
		upper := false
		opts.Upper = &upper
	}
	text, err := morsekit.DecodeMorse(code, opts)
	if err != nil {
		fmt.Fprintf(stderr, "morsekit: decode: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, text)
	return 0
}

func cmdTiming(f *flagSet, stdin io.Reader, stdout, stderr io.Writer) int {
	code, err := readInput(f, stdin)
	if err != nil {
		fmt.Fprintf(stderr, "morsekit: timing: %v\n", err)
		return 1
	}
	counts, err := morsekit.AnalyzePacing(code)
	if err != nil {
		fmt.Fprintf(stderr, "morsekit: timing: %v\n", err)
		return 1
	}
	wpm := 20.0
	if f.wpmSet {
		wpm = f.wpm
	}
	dot, dash, intra, letter, word, err := morsekit.ElementDurations(0, wpm)
	if err != nil {
		fmt.Fprintf(stderr, "morsekit: timing: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "source: %s\n", code)
	for _, key := range []string{
		morsekit.PacingSymbols, morsekit.PacingDots, morsekit.PacingDashes,
		morsekit.PacingLetters, morsekit.PacingWords,
	} {
		fmt.Fprintf(stdout, "%s: %d\n", key, counts[key])
	}
	fmt.Fprintf(stdout, "\nelement durations (PARIS standard, unit = 1200/wpm ms)\n")
	fmt.Fprintf(stdout, "wpm: %g (unit/dot = %.1f ms)\n", wpm, dot)
	fmt.Fprintf(stdout, "dash: %.1f ms\n", dash)
	fmt.Fprintf(stdout, "intra-char gap: %.1f ms\n", intra)
	fmt.Fprintf(stdout, "letter gap: %.1f ms\n", letter)
	fmt.Fprintf(stdout, "word gap: %.1f ms\n", word)
	return 0
}