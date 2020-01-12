package text

import (
	"unicode"

	"github.com/Spriithy/rosa/pkg/compiler/fragments"
)

const (
	LF = '\u000A'
	FF = '\u000C'
	CR = '\u000D'
	SU = '\u001A'
)

type Escape []rune

func DigitToInt(r rune, base int) (val int) {
	switch {
	case r < '9':
		val = int(r - '0')
	case 'a' <= r && r <= 'z':
		val = int(r - 'a' + 10)
	case 'A' <= r && r <= 'Z':
		val = int(r - 'A' + 10)
	default:
		val = -1
	}
	if !(0 <= val && val < base) {
		val = -1
	}
	return
}

func EscapeBuffer(n int) Escape {
	return make(Escape, n)
}

func (e Escape) Rune() (r rune) {
	for i, d := range e {
		r |= rune(DigitToInt(d, 16)) << ((len(e) - i - 1) * 4)
	}
	return
}

func (e Escape) String() string {
	return "\\u" + string(e)
}

func EscapeRune(r rune) string {
	hex := func(ch rune) (r rune) {
		if ch < 10 {
			r = '0'
		} else {
			r = 'A'
		}
		r += ch
		return
	}
	seq := EscapeBuffer(4)
	seq[0] = hex((r >> 12))
	seq[1] = hex((r >> 8) % 16)
	seq[2] = hex((r >> 4) % 16)
	seq[3] = hex((r >> 0) % 16)
	return string(seq)
}

func IsLineBreakRune(r rune) bool {
	switch r {
	case LF, FF, CR, SU:
		return true
	default:
		return false
	}
}

func IsWhitespace(r rune) bool {
	switch r {
	case ' ', '\t', CR:
		return true
	default:
		return false
	}
}

func Any(r rune) bool {
	return true
}

func IsVarPart(r rune) bool {
	return '0' <= r && r <= '9' || 'A' <= r && r <= 'Z' || 'a' <= r && r <= 'z'
}

func IsIdentifierStart(r rune) bool {
	return r == '_' || r == '$' || unicode.Is(unicode.Lu, r)
}

func IsIdentifierPart(r rune) bool {
	return r == '$' || unicode.Is(unicode.Lu, r)
}

func IsSpecial(r rune) bool {
	return unicode.In(r, unicode.Sm, unicode.So)
}

func IsOperatorPart(r rune) bool {
	switch r {
	case '~', '!', '@', '#', '%', '^', '*', '+', '-', '<', '>', '?', ':', '=', '&', '|', '/', '\\':
		return true
	default:
		return IsSpecial(r)
	}
}

func IsParenMatch(left, right rune) bool {
	switch left {
	case '(':
		return right == ')'
	case '[':
		return right == ']'
	case '{':
		return right == '}'
	default:
		return false
	}
}

var (
	IsSeparator  = fragments.Any('(', ')', '[', ']', '{', '}', '.', ',', ';')
	NonZeroDigit = fragments.Range('1', '9')
	Digit        = fragments.Range('0', '9')
	BinaryDigit  = fragments.Any('0', '1')
	OctalDigit   = fragments.Range('0', '7')
	HexDigit     = fragments.Or(fragments.Range('0', '9'), fragments.Range('a', 'f'), fragments.Range('A', 'F'))
	Exponent     = fragments.Any('e', 'E')
	Sign         = fragments.Any('+', '-')
	Lower        = fragments.Or(fragments.Range('a', 'z'), fragments.In(unicode.Ll))
	Upper        = fragments.Or(fragments.Range('A', 'Z'), fragments.In(unicode.Lu))
	Letter       = fragments.Or(Lower, Upper, fragments.In(unicode.Lo, unicode.Lt))
	IdentStart   = fragments.Or(fragments.Rune('$'), fragments.Rune('_'), Upper, Lower)
	IdentRest    = fragments.Or(fragments.Rune('$'), Upper, Lower, Digit)
)
