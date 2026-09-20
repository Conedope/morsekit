package morsekit

import "strings"

// ITU-R M.1677-1 international Morse code tables.
//
// Code table reference (kept in code for precision):
//
//	Letters:   A .-    B -...   C -.-.   D -..    E .      F ..-.
//	           G --.   H ....   I ..     J .---   K -.-    L .-..
//	           M --    N -.     O ---    P .--.   Q --.-   R .-.
//	           S ...   T -      U ..-    V ...-   W .--    X -..-
//	           Y -.--  Z --..
//	Digits:    0 ----- 1 .----  2 ..---  3 ...--  4 ....-  5 .....
//	           6 -.... 7 --...  8 ---..  9 ----.
//	Punctuation: see punctCodes below (ITU forms).
//	Prosigns:   see prosignCodes below.

// letterCodes maps each ASCII letter (upper case) to its Morse symbols.
var letterCodes = map[rune]string{
	'A': ".-", 'B': "-...", 'C': "-.-.", 'D': "-..", 'E': ".", 'F': "..-.",
	'G': "--.", 'H': "....", 'I': "..", 'J': ".---", 'K': "-.-", 'L': ".-..",
	'M': "--", 'N': "-.", 'O': "---", 'P': ".--.", 'Q': "--.-", 'R': ".-.",
	'S': "...", 'T': "-", 'U': "..-", 'V': "...-", 'W': ".--", 'X': "-..-",
	'Y': "-.--", 'Z': "--..",
}

// digitCodes maps ASCII digits 0-9 to their Morse symbols.
var digitCodes = map[rune]string{
	'0': "-----", '1': ".----", '2': "..---", '3': "...--", '4': "....-",
	'5': ".....", '6': "-....", '7': "--...", '8': "---..", '9': "----.",
}

// punctCodes maps common punctuation to its ITU Morse form:
//
//	.  Period       .-.-.-     (  Open paren   -.--.
//	,  Comma        --..--     )  Close paren  -.--.-
//	?  Question     ..--..     -  Hyphen       -....-
//	'  Apostrophe   .----.     /  Slash        -..-.
//	!  Exclamation  -.-.--     @  At           .--.-.
var punctCodes = map[rune]string{
	'.': ".-.-.-", ',': "--..--", '?': "..--..", '\'': ".----.",
	'!': "-.-.--", '/': "-..-.", '@': ".--.-.", '(': "-.--.", ')': "-.--.-",
	'-': "-....-",
}

// prosignCodes maps prosign tokens (as written in plain text) to symbols:
//
//	<AR>  End of message   .-.-.
//	<AS>  Wait             .-...
//	<BT>  Break            -...-
//	<KN>  Invite reply      -.--.   (identical to '(' by ITU convention)
//	<SK>  End of work       ...-.-
//	<HH>  Error             ........
var prosignCodes = map[string]string{
	"<AR>": ".-.-.", "<AS>": ".-...", "<BT>": "-...-",
	"<KN>": "-.--.", "<SK>": "...-.-", "<HH>": "........",
}

// decodeTable maps a Morse token back to its plain-text form. Letters, digits
// and punctuation are inserted first, so when a prosign shares a code with
// punctuation (the genuine ITU overlap between <KN> and "(" - both "-.--.")
// the punctuation wins and plain text round-trips exactly. <KN> still encodes
// to the same symbols.
var decodeTable = buildDecodeTable()

func buildDecodeTable() map[string]string {
	t := make(map[string]string, len(letterCodes)+len(digitCodes)+len(punctCodes)+len(prosignCodes))
	add := func(table map[rune]string) {
		for ch, code := range table {
			t[code] = string(ch)
		}
	}
	add(letterCodes)
	add(digitCodes)
	add(punctCodes)
	for tok, code := range prosignCodes {
		if _, taken := t[code]; !taken {
			t[code] = tok
		}
	}
	return t
}

// TableChars returns every single text character the tables can encode
// (letters A-Z, digits 0-9 and the punctuation below; prosigns are tokens,
// not characters, and are excluded).
func TableChars() string {
	var b strings.Builder
	for ch := 'A'; ch <= 'Z'; ch++ {
		b.WriteRune(ch)
	}
	for ch := '0'; ch <= '9'; ch++ {
		b.WriteRune(ch)
	}
	for _, ch := range []rune{'.', ',', '?', '\'', '!', '/', '@', '(', ')', '-'} {
		b.WriteRune(ch)
	}
	return b.String()
}