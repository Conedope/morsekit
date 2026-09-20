# morsekit

An offline Morse code toolkit for Go (ITU-R M.1677-1). Text to Morse, Morse
to text, prosigns, and PARIS-standard timing-unit analysis. Standard library
only, no network access.

## Features

- `EncodeText` — text → Morse. ASCII letters (case-folded), digits, ITU
  punctuation, and inline prosign tokens like `<AR>`.
- `DecodeMorse` — Morse → text with word/letter spacing, uppercase (default)
  or lowercase output, custom word separator, and unknown sequences either
  flagged as errors or replaced with `?` (`IgnoreUnknown`).
- `AnalyzePacing` — per-message counts: symbols, dots, dashes, letters, words.
- `ElementDurations` — dot/dash/gap timings in milliseconds for a given WPM,
  using the PARIS formula (unit = 1200/wpm ms).
- Thin CLI: `encode`, `decode`, `timing`, `--version`, `--help`.

## Install / build

```sh
go build ./cmd/morsekit        # produces ./morsekit
# or: go install github.com/Conedope/morsekit/cmd/morsekit@latest
```

Requires Go 1.22+. Build with `CGO_ENABLED=0` for a static binary.

## CLI usage

```
morsekit encode [flags] [TEXT]    text -> Morse code
morsekit decode [flags] [CODE]    Morse code -> text (uppercase by default)
morsekit timing [flags] [CODE]    symbol counts + element durations (PARIS)

  -f, --file FILE       read input from FILE instead of arguments/stdin
      --lower           decode: output lowercase text
      --ignore-unknown  decode: replace unknown tokens with "?" and exit 0
      --wpm N           timing: words per minute (default 20)
  -h, --help | --version
```

Input comes from the arguments (joined with spaces), else `--file`, else
stdin. Exit status: `0` success, `1` data error, `2` usage error. Encoding an
empty input prints an empty line and exits `0`.

### Verified samples (real output)

```
$ morsekit encode "MORSE CODE TOOLKIT"
-- --- .-. ... . / -.-. --- -.. . / - --- --- .-.. -.- .. -

$ morsekit decode ".... . .-.. .-.. --- / .-- --- .-. .-.. -.."
HELLO WORLD

$ morsekit timing ".... . .-.. .-.. --- / .-- --- .-. .-.. -.." --wpm 20
source: .... . .-.. .-.. --- / .-- --- .-. .-.. -..
symbols: 32
dots: 19
dashes: 13
letters: 10
words: 2

element durations (PARIS standard, unit = 1200/wpm ms)
wpm: 20 (unit/dot = 60.0 ms)
dash: 180.0 ms
intra-char gap: 60.0 ms
letter gap: 180.0 ms
word gap: 420.0 ms

$ morsekit encode "SOS CQ <AR>"
... --- ... / -.-. --.- / .-.-.

$ morsekit --version
morsekit 0.1.0
```

## Library

```go
code, _ := morsekit.EncodeText("Hello World")
// ".... . .-.. .-.. --- / .-- --- .-. .-.. -.."

text, _ := morsekit.DecodeMorse(".... . .-.. .-.. --- / .-- --- .-. .-.. -..", morsekit.DecodeOpts{})
// "HELLO WORLD"

text, _ = morsekit.DecodeMorse(".... . .-.. .-.. ---", morsekit.DecodeOpts{Upper: ptr(false)})
// "hello"

round, _ := morsekit.Roundtrip("Hello World")   // "HELLO WORLD" (upper-normalised)

counts, _ := morsekit.AnalyzePacing(".... . / .-- .")
// {symbols:7, dots:4, dashes:3, letters:4, words:2}

dot, dash, intra, letter, word, _ := morsekit.ElementDurations(0, 20)
// 60, 180, 60, 180, 420  (milliseconds)
```

Decisions worth knowing:

- `DecodeOpts` zero value is uppercase output, `" "` word separator, unknown
  tokens rejected. `--lower`/`Upper: &false` switch to lowercase. Unknown
  tokens return an `*UnknownTokenError` (wrapping `ErrUnknownProsign`) unless
  `IgnoreUnknown` is set, in which case each is replaced by `?`.
- `EncodeText` errors (wrapping `ErrUnsupportedChar`) on anything outside the
  tables: non-ASCII letters, punctuation without a table entry, and malformed
  or unknown `<...>` prosign tokens. The error names the character.
- `<KN>` and `(` share the code `-.--.` (a genuine ITU overlap). Decode
  resolves that code to `(` so plain text round-trips exactly; `<KN>`
  encodes to the same symbols.

## Code tables

Letters A-Z and digits 0-9 follow ITU-R M.1677-1 (A `.-`, B `-...`, … Z `--..`;
0 `-----`, … 9 `----.`). Punctuation:

| Char  | Morse    | Name          | Char | Morse    | Name       |
|-------|----------|---------------|------|----------|------------|
| `.`   | `.-.-.-` | period        | `(`  | `-.--.`  | left paren |
| `,`   | `--..--` | comma         | `)`  | `-.--.-` | right paren|
| `?`   | `..--..` | question      | `-`  | `-....-` | hyphen     |
| `'`   | `.----.` | apostrophe    | `/`  | `-..-.`  | slash      |
| `!`   | `-.-.--` | exclamation   | `@`  | `.--.-.` | at         |

### Prosigns

| Token | Morse      | Meaning        |
|-------|------------|----------------|
| `<AR>`| `.-.-.`    | end of message |
| `<AS>`| `.-...`    | wait           |
| `<BT>`| `-...-`    | break          |
| `<KN>`| `-.--.`    | invite reply   |
| `<SK>`| `...-.-`   | end of work    |
| `<HH>`| `........` | error          |

## Timing math

PARIS standard: the word `PARIS` plus its trailing word gap totals **50
units**. Element lengths: dot = 1 unit, dash = 3, intra-character gap = 1,
letter gap = 3, word gap = 7.

```
unit (ms) = 1200 / wpm
dot       = 1u   = 1200/wpm ms        e.g. 20 wpm ->   60 ms
dash      = 3u   = 3600/wpm ms        e.g. 20 wpm ->  180 ms
letter gap= 3u   = 3600/wpm ms        e.g. 20 wpm ->  180 ms
word gap  = 7u   = 8400/wpm ms        e.g. 20 wpm ->  420 ms
```

`ElementDurations(units, wpm)` accepts an optional base-unit override in
milliseconds (pass `0` to derive from `wpm`); `wpm <= 0` returns
`ErrBadWPM`.

## Test

```sh
go vet ./...
go test ./...
```

CI (`.github/workflows/ci.yml`) runs `setup-go@v5` then `go vet ./...` and
`go test ./...` on push and pull requests.

## License

MIT, © 2026 Conedope.