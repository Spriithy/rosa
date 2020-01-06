package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Pos struct {
	Line int
	Col  int
}

type Scanner struct {
	path        string
	source      []byte
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
		source: source,
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

func (s *Scanner) peek() byte {
	if s.eof() {
		return 0
	}
	return s.source[s.current]
}

func (s *Scanner) advance() byte {
	if s.peek() == '\n' {
		s.lastNewline = s.current + 1
		s.line++
	}
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) match(expected ...byte) bool {
	if s.eof() {
		return false
	}
	for _, b := range expected {
		if s.peek() == b {
			s.advance()
			return true
		}
	}
	return false
}

func (s *Scanner) matchIf(fs ...func(byte) bool) bool {
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
	case s.match(' ', '\n', '\r', '\t'):
		token = s.next()
	case s.match('/'):
		if s.match('/') {
			s.comment()
			token = s.next()
		} else {
			token = Token{
				Type: Identifier,
				Pos:  s.pos(),
				Text: s.text(),
			}
		}
	case s.match('(', ')', '[', ']', '{', '}', ',', ';'):
		token = Token{
			Type: Separator,
			Pos:  s.pos(),
			Text: s.text(),
		}
	case s.match('"'):
		token = s.string()
	case s.matchIf(isAlphaIdentStart):
		token = s.alphaIdent()
	case s.matchIf(isOpIdentPart):
		token = s.opIdent()
	case s.matchIf(isDigit):
		digit := s.text()[0]
		if digit == '0' {
			switch {
			case s.match('b', 'B'):
				token = s.number(2, true)
			case s.match('x', 'X'):
				token = s.number(16, true)
			default:
				token = s.number(10, false)
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

func isAlphaIdentStart(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || b == '_'
}

func isAlphaIdentPart(b byte) bool {
	return isAlphaIdentStart(b) || (b >= '0' && b <= '9')
}

func isOpIdentPart(b byte) bool {
	return strings.IndexByte("@$%<>+-*/:=~#&|^!?", b) > 0
}

const (
	digits      = "0123456789abcdefghijklmnopqrstuvwxyz"
	digitsUpper = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

func isDigitBase(base int) func(byte) bool {
	return func(b byte) bool {
		index := strings.IndexByte(digits, b)
		indexUpper := strings.IndexByte(digitsUpper, b)
		return (index > 0 && index < base) || (indexUpper > 0 && indexUpper < base)
	}
}

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
	isDigit = isDigitBase(10)
	isHex   = isDigitBase(16)

	digitValue = digitBaseValue(10)
	hexValue   = digitBaseValue(16)
)

////////////////////////////////////////////////////////////////////////////////

func (s *Scanner) comment() {
	for s.peek() != '\n' && !s.eof() {
		s.advance()
	}
}

func (s *Scanner) number(base int, atLeastOne bool) (token Token) {
	decimal := false
	if atLeastOne && !s.matchIf(isDigitBase(base)) {
		s.error(s.pos(), "expected at least one base %d digit after '%s'", base, s.text())
		token = Token{
			Type: Error,
			Pos:  s.pos(),
			Text: s.text(),
		}
		return
	}
	for s.matchIf(isDigitBase(base)) {
		// continue scanning
	}
	if base == 10 && s.match('.') {
		decimal = true
		// decimal number
		for s.matchIf(isDigit) {
			// continue scanning
		}
	}
	if base == 10 && s.match('e', 'E') {
		// scan exponent
		s.match('+', '-')
		for s.matchIf(isDigit) {
			// continue scanning
		}
	}
	typ := Integer
	if decimal {
		typ = Float
	}
	token = Token{
		Type: typ,
		Pos:  s.pos(),
		Text: s.text(),
	}
	return
}

func (s *Scanner) alphaIdent() (token Token) {
	for s.matchIf(isAlphaIdentPart) {
		// continue scanning
	}
	typ, ident := Identifier, s.text()
	if exists(keywords)(ident) {
		typ = Keyword
	}
	token = Token{
		Type: typ,
		Pos:  s.pos(),
		Text: s.text(),
	}
	return
}

func (s *Scanner) opIdent() (token Token) {
	for s.matchIf(isOpIdentPart) {
		// continue scanning
	}
	typ, ident := OperatorIdentifier, s.text()
	if exists(operators)(ident) {
		typ = Operator
	} else if exists(separators)(ident) {
		typ = Separator
	}
	token = Token{
		Type: typ,
		Pos:  s.pos(),
		Text: s.text(),
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
		first := s.matchIf(isHex)
		second := s.matchIf(isHex)
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
