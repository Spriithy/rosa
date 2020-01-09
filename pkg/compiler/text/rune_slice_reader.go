package text

import (
	"fmt"
	"io/ioutil"
	"os"
)

type RuneSliceReader struct {
	buf []rune

	// The last rune read
	r rune

	// The offset one past the last read rune
	runeOffset int

	// The start offset of the current line
	lineStartOffset int

	// The start offset of the line before the current one
	lastLineStartOffset int

	// The last index at which an unicode escape was found
	lastUnicodeOffset int

	decodeUni func() bool

	// An error routine to call on bad unicode escapes
	error func(int, string)

	lookahead bool
}

func FileRuneSliceReader(path string, error func(int, string)) (*RuneSliceReader, error) {
	fileExists := func(path string) bool {
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return false
		}
		return !info.IsDir()
	}
	if !fileExists(path) {
		return nil, fmt.Errorf("error: %s doesn't exist", path)
	}
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error: failed to open %s. %s", path, err.Error())
	}
	rsc := NewRuneSliceReader(error)
	rsc.buf = []rune(string(source))
	return rsc, nil
}

func NewRuneSliceReader(error func(int, string)) *RuneSliceReader {
	return &RuneSliceReader{
		lastUnicodeOffset: -1,
		decodeUni:         func() bool { return true },
		error:             error,
	}
}

func CopyRuneSliceReader(other *RuneSliceReader, lookahead bool) *RuneSliceReader {
	return &RuneSliceReader{
		buf:                 other.buf,
		r:                   other.r,
		runeOffset:          other.runeOffset,
		lineStartOffset:     other.lineStartOffset,
		lastLineStartOffset: other.lastLineStartOffset,
		lastUnicodeOffset:   other.lastUnicodeOffset,
		decodeUni:           other.decodeUni,
		error:               other.error,
		lookahead:           lookahead,
	}
}

func LookaheadRuneSliceReader(other *RuneSliceReader) *RuneSliceReader {
	return CopyRuneSliceReader(other, true)
}

// Is last character a unicode escape ?
func (rsc *RuneSliceReader) IsUnicodeEscape() bool {
	return rsc.runeOffset == rsc.lastUnicodeOffset
}

// Advance one rune. Reducing CR;LF pairs to just LF
func (rsc *RuneSliceReader) NextRune() rune {
	if rsc.runeOffset >= len(rsc.buf) {
		rsc.r = SU
	} else {
		rsc.r = rsc.buf[rsc.runeOffset]
		rsc.runeOffset++
		if rsc.r == '\\' {
			rsc.potentialUnicode()
		}
		if rsc.r < ' ' {
			rsc.skipCR()
			rsc.potentialLineEnd()
		}
	}
	return rsc.r
}

// Advance one rune but don't reduc CR;LF pairs
func (rsc *RuneSliceReader) NextRawRune() {
	if rsc.runeOffset >= len(rsc.buf) {
		rsc.r = SU
	} else {
		rsc.r = rsc.buf[rsc.runeOffset]
		rsc.runeOffset++
		if rsc.r == '\\' {
			rsc.potentialUnicode()
		}
	}
}

// Interpret \\uXXXX escape sequences
func (rsc *RuneSliceReader) potentialUnicode() {
	evenSlashPrefix := func() bool {
		p := rsc.runeOffset - 2
		for p >= 0 && rsc.buf[p] == '\\' {
			p--
		}
		return (rsc.runeOffset - p%2) == 0
	}

	udigit := func() rune {
		if rsc.runeOffset >= len(rsc.buf) {
			rsc.error(rsc.runeOffset-1, "incomplete unicode escape")
			return SU
		}
		d := DigitToInt(rsc.buf[rsc.runeOffset], 16)
		if d > 0 {
			rsc.runeOffset++
		} else {
			rsc.error(rsc.runeOffset, "error in unicode escape")
		}
		return rune(d)
	}

	if rsc.runeOffset < len(rsc.buf) && rsc.buf[rsc.runeOffset] == 'u' && rsc.decodeUni() && evenSlashPrefix() {
		for ok := true; ok; ok = rsc.runeOffset < len(rsc.buf) && rsc.buf[rsc.runeOffset] == 'u' {
			rsc.runeOffset++
		}
		var code rune
		code |= udigit() << 12
		code |= udigit() << 8
		code |= udigit() << 4
		code |= udigit()
		rsc.lastUnicodeOffset = rsc.runeOffset
		rsc.r = code
	}
}

func (rsc *RuneSliceReader) skipCR() {
	if rsc.r == CR && rsc.runeOffset < len(rsc.buf) {
		switch rsc.buf[rsc.runeOffset] {
		case LF:
			rsc.runeOffset++
			rsc.r = LF
		case '\\':
			lookahead := CopyRuneSliceReader(rsc, true)
			if lookahead.GetU() == LF {
				rsc.potentialUnicode()
			}
		}
	}
}

func (rsc *RuneSliceReader) potentialLineEnd() {
	if rsc.r == LF || rsc.r == FF {
		rsc.lastLineStartOffset = rsc.lineStartOffset
		rsc.lineStartOffset = rsc.runeOffset
	}
}

// Lookahead RuneSliceReader only-methods

func (rsc *RuneSliceReader) GetC() rune {
	if !rsc.lookahead {
		panic("RuneSliceReader.GetC() call from non-lookahead RuneSliceReader")
	}
	return rsc.NextRune()
}

func (rsc *RuneSliceReader) GetU() rune {
	if !rsc.lookahead {
		panic("RuneSliceReader.GetU() call from non-lookahead RuneSliceReader")
	}
	if rsc.buf[rsc.runeOffset] != '\\' {
		panic("RuneSliceReader.Get() requires to match a backquote ('\\') beforehand")
	}
	rsc.r = '\\'
	rsc.runeOffset++
	rsc.potentialUnicode()
	return rsc.r
}
