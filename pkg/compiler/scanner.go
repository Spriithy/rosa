package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"unicode"

	"github.com/Spriithy/rosa/pkg/compiler/fragments"
)

type Pos struct {
	Line int
	Col  int
}

type Scanner struct {
	path        string
	source      []rune
	tokens      []Token
	start       int
	current     int
	line        int
	lastNewline int
	Logs        []Log
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func NewScanner(path string) (scanner *Scanner) {
	if !fileExists(path) {
		fmt.Printf("error: %s doesn't exist\n", path)
		return
	}
	source, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("error: failed to open %s\n", path)
		return
	}
	scanner = &Scanner{
		path:   path,
		line:   1,
		source: []rune(string(source)),
	}
	return
}

////////////////////////////////////////////////////////////////////////////////

func (s *Scanner) Scan() (token Token) {
	if s.eof() {
		token = Token{
			Type: EOF,
			Pos:  s.currentPos(),
		}
		return
	}
	token = s.next()
	s.tokens = append(s.tokens, token)
	return
}

func (s *Scanner) error(pos Pos, message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	message = fmt.Sprintf("%s:%d:%d: %s", s.path, pos.Line, pos.Col, message)
	s.Logs = append(s.Logs, Log{
		Level:   LogError,
		Pos:     pos,
		Message: message,
	})
}

func (s *Scanner) warning(pos Pos, message string, args ...interface{}) {
	message = fmt.Sprintf(message, args...)
	message = fmt.Sprintf("%s:%d:%d: %s", s.path, pos.Line, pos.Col, message)
	s.Logs = append(s.Logs, Log{
		Level:   LogWarning,
		Pos:     pos,
		Message: message,
	})
}

func (s *Scanner) eof() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) col() int {
	return s.start - s.lastNewline + 1
}

func (s *Scanner) currentCol() int {
	return s.current - s.lastNewline + 1
}

func (s *Scanner) pos() Pos {
	return Pos{
		Line: s.line,
		Col:  s.col(),
	}
}

func (s *Scanner) currentPos() Pos {
	return Pos{
		Line: s.line,
		Col:  s.currentCol(),
	}
}

func (s *Scanner) peek() rune {
	if s.eof() {
		return 0
	}
	return s.source[s.current]
}

func (s *Scanner) advance() rune {
	if s.peek() == '\n' {
		s.lastNewline = s.current + 1
		s.line++
	}
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) match(expected ...rune) bool {
	if s.eof() {
		return false
	}
	for _, r := range expected {
		if s.peek() == r {
			s.advance()
			return true
		}
	}
	return false
}

func (s *Scanner) matchIf(fs ...fragments.Fragment) bool {
	if s.eof() {
		return false
	}
	for _, f := range fs {
		if f(s.peek()) {
			s.advance()
			return true
		}
	}
	return false
}

func (s *Scanner) many(f fragments.Fragment) bool {
	for s.matchIf(f) {
	}
	return true
}

func (s *Scanner) atLeastOne(f fragments.Fragment) bool {
	if s.matchIf(f) {
		for s.matchIf(f) {
		}
		return true
	}
	return false
}

func (s *Scanner) text() string {
	return string(s.source[s.start:s.current])
}

func (s *Scanner) next() (token Token) {
	s.start = s.current // reset token pos
	switch {
	case s.eof():
		token = Token{
			Type: EOF,
			Pos:  s.pos(),
			Text: s.text(),
		}
	case s.matchIf(unicode.IsSpace):
		token = s.next()
	case s.match('/'):
		if s.match('/') {
			s.comment()
			token = s.next()
		} else {
			s.op()
			text := s.text()
			token = Token{
				Type: tokenType(text),
				Pos:  s.pos(),
				Text: text,
			}
		}
	case s.matchIf(sepChar):
		token = Token{
			Type: Separator,
			Pos:  s.pos(),
			Text: s.text(),
		}
	case s.match('"'):
		token = s.string()
	case s.op():
		text := s.text()
		token = Token{
			Type: tokenType(text),
			Pos:  s.pos(),
			Text: text,
		}
	case s.matchIf(letter):
		token = s.varIdent()
	case s.atLeastOne(nonZeroDigit):
		token = s.number()
	case s.match('0'):
		switch {
		case s.match('b', 'B'): // binary literal
			token = s.binary()
		case s.match('o', 'O'): // octal literal
			token = s.octal()
		case s.match('x', 'X'): // hex literal
			token = s.hex()
		case s.match('.'):
			token = s.decimalPart() // float 0.xxx
		case s.matchIf(nonZeroDigit):
			s.error(s.pos(), "numbers should not start with a '0' (use '123' instead of '0123')")
			s.many(digit)
			token = s.number()
		default:
			token = Token{
				Type: Integer,
				Pos:  s.pos(),
				Text: s.text(),
			}
		}
	default:
		s.advance()
		token = Token{
			Type: Error,
			Pos:  s.pos(),
			Text: s.text(),
		}
	}
	return
}

////////////////////////////////////////////////////////////////////////////////

/*
When an expression uses multiple operators, the operators are evaluated based on the priority of the first character:

(characters not shown below)
* / %
+ -
:
= !
< >
&
^
|
(all letters)
*/

var (
	charNoBackQuoteOrNewline = fragments.Or(
		fragments.Range('\u0020', '\u0026'),
		fragments.Range('\u0028', '\u007E'),
	)
	sepChar      = fragments.Any('(', ')', '[', ']', '{', '}', '.', ',', ';')
	opChar       = fragments.Any('!', '#', '%', '&', '*', '+', '-', ':', '<', '=', '>', '?', '@', '^', '|', '\\', '~', '$')
	nonZeroDigit = fragments.Range('1', '9')
	digit        = fragments.Range('0', '9')
	binaryDigit  = fragments.Any('0', '1')
	octalDigit   = fragments.Range('0', '7')
	hexDigit     = fragments.Or(
		fragments.Range('0', '9'),
		fragments.Range('a', 'f'),
		fragments.Range('A', 'F'),
	)
	exponentChar = fragments.Any('e', 'E')
	lower        = fragments.Or(
		fragments.Range('a', 'z'),
		fragments.In(unicode.Ll),
	)
	upper = fragments.Or(
		fragments.Range('A', 'Z'),
		fragments.Rune('$'),
		fragments.In(unicode.Lu),
	)
	letter = fragments.Or(
		lower, upper,
		fragments.In(unicode.Lo),
		fragments.In(unicode.Lt),
	)
)

const (
	digits      = "0123456789abcdefghijklmnopqrstuvwxyz"
	digitsUpper = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func digitBaseValue(base int) func(byte) int {
	return func(b byte) int {
		var val int
		if val = strings.IndexByte(digits, b); val == -1 {
			val = strings.IndexByte(digitsUpper, b)
		}
		return val % base
	}
}

var (
	binaryValue = digitBaseValue(2)
	octalValue  = digitBaseValue(8)
	digitValue  = digitBaseValue(10)
	hexValue    = digitBaseValue(16)
)

////////////////////////////////////////////////////////////////////////////////

func (s *Scanner) comment() {
	for s.peek() != '\n' && !s.eof() {
		s.advance()
	}
}

func (s *Scanner) varIdent() (token Token) {
	return s.idRest()
}

func (s *Scanner) idRest() (token Token) {
	s.many(fragments.Or(letter, digit))
	s.match('_')
	s.op()
	text := s.text()
	token = Token{
		Type: tokenType(text),
		Pos:  s.pos(),
		Text: text,
	}
	return
}

func (s *Scanner) op() (ok bool) {
	ok = s.match('/')
	ok = s.atLeastOne(opChar) || ok
	return
}

func (s *Scanner) base(digits fragments.Fragment, baseName string) (token Token) {
	if !s.atLeastOne(digit) {
		s.error(s.currentPos(), "expected at least one digit in %s integer literal", baseName)
		token = Token{
			Type: Integer,
			Pos:  s.pos(),
			Text: s.text(),
		}
		return
	}
	text := s.text()
	if offset := strings.IndexFunc(text[2:], fragments.Not(digits)); offset >= 0 {
		pos := s.pos()
		pos.Col += offset
		s.error(pos, "unexpected digit in %s literal: '%c'", baseName, text[2:][offset])
	}
	token = Token{
		Type: Integer,
		Pos:  s.pos(),
		Text: text,
	}
	return
}

func (s *Scanner) binary() (token Token) {
	return s.base(binaryDigit, "binary")
}

func (s *Scanner) octal() (token Token) {
	return s.base(octalDigit, "octal")
}

func (s *Scanner) hex() (token Token) {
	return s.base(hexDigit, "hexadecimal")
}

func (s *Scanner) tryExponent() bool {
	return exponentChar(s.peek())
}

// ('e' | 'E') ('+' | '-')? digit+
func (s *Scanner) exponent() (token Token) {
	if s.matchIf(exponentChar) {
		s.match('+', '-') // optional
		if !s.atLeastOne(digit) {
			s.error(s.currentPos(), "expected at least one exponent digit in float literal")
			text := s.text()
			cut := strings.LastIndexFunc(text, exponentChar)
			token = Token{
				Type: Float,
				Pos:  s.pos(),
				Text: text[:cut],
			}
			return
		}
		token = Token{
			Type: Float,
			Pos:  s.pos(),
			Text: s.text(),
		}
		return
	}
	token = Token{
		Type: Float,
		Pos:  s.pos(),
		Text: s.text(),
	}
	return
}

// digit+ exponent?
func (s *Scanner) decimalPart() (token Token) {
	if !s.atLeastOne(digit) {
		s.error(s.currentPos(), "expected at least one digit after decimal point in float literal")
	}
	token = s.exponent()
	return
}

// digit+ ('.' decimalPart | exponent)?
func (s *Scanner) number() (token Token) {
	switch {
	case s.match('.'):
		token = s.decimalPart()
	case s.tryExponent():
		token = s.exponent()
	default:
		token = Token{
			Type: Integer,
			Pos:  s.pos(),
			Text: s.text(),
		}
	}
	return
}

func (s *Scanner) escape() (seq string) {
	pos := Pos{s.line, s.currentCol() - 2}
	switch {
	case s.match('n'):
		seq = "\n"
	case s.match('r'):
		seq = "\r"
	case s.match('t'):
		seq = "\t"
	case s.match('\\'):
		seq = "\\"
	case s.match('"'):
		seq = "\""
	case s.match('x'):
		first := s.matchIf(hexDigit)
		second := s.matchIf(hexDigit)
		if !first || !second {
			s.error(pos, "malformed hexadecimal escape sequence. Use '\\x1f' for instance.")
		}
		text := s.text()
		val := 16*hexValue(text[len(text)-2]) + hexValue(text[len(text)-1])
		if val == 0 {
			s.warning(pos, "null bytes in string literals are ignored")
			return
		}
		seq = string(val)
	case s.match('0'):
		s.warning(pos, "null bytes in string literals are ignored")
	}
	return
}

func (s *Scanner) string() (token Token) {
	var value string
	for {
		switch {
		case s.eof() || s.match('\n'):
			s.error(s.pos(), "unclosed string literal")
			token = Token{
				Type: Error,
				Pos:  s.pos(),
				Text: s.text(),
			}
			return
		case s.match('\\'):
			value += s.escape()
		case s.match('"'):
			token = Token{
				Type: String,
				Pos:  s.pos(),
				Text: value,
			}
			return
		default:
			value += string(s.advance())
		}
	}
}
