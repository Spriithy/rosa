package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/Spriithy/rosa/pkg/compiler/fragments"
	"github.com/Spriithy/rosa/pkg/compiler/text"
)

type Scanner struct {
	path        string
	source      []rune
	tokens      []text.Token
	start       int
	current     int
	line        int
	lastNewline int
	Logs        []Log

	openComments int
	parens       stack
	tokenData    strings.Builder
}

type tokenStack []text.Token

type stack interface {
	push(text.Token)
	peek() text.Token
	pop() text.Token
	isEmpty() bool
}

func (s *tokenStack) push(token text.Token) {
	*s = append(*s, token)
}

func (s *tokenStack) pop() (token text.Token) {
	if s.isEmpty() {
		return
	}
	last := len(*s) - 1
	token = (*s)[last]
	*s = (*s)[:last]
	return
}

func (s *tokenStack) peek() (token text.Token) {
	if s.isEmpty() {
		return
	}
	last := len(*s) - 1
	token = (*s)[last]
	return
}

func (s *tokenStack) isEmpty() bool {
	return len(*s) == 0
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
		parens: new(tokenStack),
	}
	return
}

////////////////////////////////////////////////////////////////////////////////

func (s *Scanner) Scan() (token text.Token) {
	if s.eof() {
		token = text.Token{
			Type: text.EOF,
			Pos:  s.currentPos(),
		}
		return
	}
	token = s.next()
	s.tokens = append(s.tokens, token)
	return
}

func (s *Scanner) error(pos text.Pos, message string, args ...interface{}) {
	s.Logs = append(s.Logs, Log{
		Path:    s.path,
		Level:   LogError,
		Pos:     pos,
		Message: fmt.Sprintf(message, args...),
	})
}

func (s *Scanner) syntaxError(pos text.Pos, message string, args ...interface{}) {
	s.Logs = append(s.Logs, Log{
		Path:    s.path,
		Level:   LogSyntaxError,
		Pos:     pos,
		Message: fmt.Sprintf(message, args...),
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

func (s *Scanner) pos() text.Pos {
	return text.Pos{
		FileName: s.path,
		Line:     s.line,
		Column:   s.col(),
	}
}

func (s *Scanner) currentPos() text.Pos {
	return text.Pos{
		FileName: s.path,
		Line:     s.line,
		Column:   s.currentCol(),
	}
}

func (s *Scanner) peek() rune {
	if s.eof() {
		return text.SU
	}
	return s.source[s.current]
}

func (s *Scanner) advance() rune {
	if s.peek() == '\n' {
		s.lastNewline = s.current + 1
		s.line++
	}
	s.ingest(s.source[s.current])
	s.current++
	return s.source[s.current-1]
}

func (s *Scanner) skipRune() {
	if s.peek() == '\n' {
		s.lastNewline = s.current + 1
		s.line++
	}
	s.current++
}

func (s *Scanner) accept(expected ...rune) bool {
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

func (s *Scanner) acceptIf(fs ...fragments.Fragment) bool {
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

func (s *Scanner) match(accepted ...rune) bool {
	if s.eof() {
		return false
	}
	for _, r := range accepted {
		if s.peek() == r {
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
			return true
		}
	}
	return false
}

func (s *Scanner) many(f fragments.Fragment) bool {
	for s.acceptIf(f) {
	}
	return true
}

func (s *Scanner) atLeastOne(f fragments.Fragment) bool {
	if s.acceptIf(f) {
		for s.acceptIf(f) {
		}
		return true
	}
	return false
}

func (s *Scanner) text() string {
	return string(s.source[s.start:s.current])
}

func (s *Scanner) data() string {
	return s.tokenData.String()
}

func (s *Scanner) ingest(r rune) {
	s.tokenData.WriteRune(r)
}

func (s *Scanner) tokenType() text.TokenType {
	return text.TypeOfToken(s.data())
}

func (s *Scanner) wrapToken() text.Token {
	return text.Token{
		Text:  s.data(),
		Type:  s.tokenType(),
		Pos:   s.pos(),
		Spans: s.start + len(s.text()),
	}
}

func (s *Scanner) wrapTokenAs(typ text.TokenType) text.Token {
	return text.Token{
		Text:  s.data(),
		Type:  typ,
		Pos:   s.pos(),
		Spans: s.start + len(s.text()),
	}
}

func (s *Scanner) wrapTokenWith(typ text.TokenType, data string) text.Token {
	return text.Token{
		Text:  data,
		Type:  typ,
		Pos:   s.pos(),
		Spans: s.start + len(s.text()),
	}
}

func (s *Scanner) next() (token text.Token) {
	s.start = s.current // reset token pos
	switch {
	case s.eof():
		token = s.wrapTokenWith(text.EOF, s.text())
	case s.match(' ', '\t', text.CR, text.LF, text.FF):
		s.skipRune()
		token = s.next()
	case s.acceptIf(text.IdentStart):
		s.identRest()
		token = s.wrapToken()
	case s.match('/'):
		s.skipRune()
		if s.skipComment() {
			token = s.next()
		} else {
			s.ingest('/')
			s.operatorRest()
			token = s.wrapToken()
		}
	case s.acceptIf(text.IsOperatorPart):
		s.operatorRest()
		token = s.wrapToken()
	case s.accept('0'):
		switch {
		case s.accept('b', 'B'):
			token = s.binary()
		case s.accept('x', 'X'):
			token = s.hexadecimal()
		default:
			token = s.number()
		}
		s.number()
		token = s.wrapTokenAs(text.IntegerLit)
	case s.acceptIf(text.NonZeroDigit):
		token = s.number()
	case s.acceptIf(text.IsSeparator):
		token = s.wrapToken()
		switch {
		case text.Lpar(token), text.Lbrk(token), text.Lbrc(token):
			s.parens.push(token)
		case text.Rpar(token), text.Rbrk(token), text.Rbrc(token):
			switch {
			case s.parens.isEmpty():
				s.syntaxError(s.pos(), "%s unexpected", token.Type)
				token = s.next()
			case text.IsParenMatch(rune(s.parens.peek().Text[0]), rune(token.Text[0])):
				s.parens.pop()
			default:
				s.syntaxError(s.pos(), "%s", s.parens)
				token = s.next()
			}
		}
	case s.match('"'):
		s.skipRune()
		if s.match('"') {
			s.skipRune()
			if s.match('"') {
				// s.rawString()
			}
		} else {
			s.stringLit()
		}
		token = s.wrapTokenAs(text.StringLit)
	case s.match('\''):
		s.charLit()
		token = s.wrapTokenAs(text.CharLit)
	default:
		s.advance()
		token = s.wrapTokenWith(text.ErrorType, s.text())
	}
	s.tokenData.Reset()
	return
}

////////////////////////////////////////////////////////////////////////////////
// Comments

func (s *Scanner) skipComment() bool {
	switch ch := s.peek(); {
	case s.match('/', '*'):
		s.skipRune()
		s.skipCommentToEnd(ch == '/')
		return true
	}
	return false
}

func (s *Scanner) skipCommentToEnd(isLineComment bool) {
	if isLineComment {
		s.skipLineComment()
	} else {
		s.openComments = 1
		s.skipNestedComments()
	}
}

func (s *Scanner) skipLineComment() {
	for !s.match(text.SU, text.CR, text.LF) {
		s.skipRune()
	}
}

func (s *Scanner) skipNestedComments() {
	switch s.peek() {
	case '/':
		s.maybeOpen()
		s.skipNestedComments()
	case '*':
		if !s.maybeClose() {
			s.skipNestedComments()
		}
	case text.SU:
		s.error(s.currentPos(), "unclosed multiline comment")
	default:
		s.skipRune()
		s.skipNestedComments()
	}
}

func (s *Scanner) maybeOpen() {
	s.skipRune()
	if s.match('*') {
		s.skipRune()
		s.openComments++
	}
}

func (s *Scanner) maybeClose() bool {
	s.skipRune()
	if s.match('/') {
		s.skipRune()
		s.openComments--
	}
	return s.openComments == 0
}

////////////////////////////////////////////////////////////////////////////////
// Identifiers & Operators

func (s *Scanner) identRest() {
	switch {
	case s.acceptIf(text.IdentRest):
		s.identRest()
	case s.accept('_'):
		s.identOrOperatorRest()
	case s.acceptIf(text.IsIdentifierPart):
		s.identRest()
	}
}

func (s *Scanner) identOrOperatorRest() {
	switch {
	case s.matchIf(text.IsIdentifierPart):
		s.identRest()
	case s.matchIf(text.IsOperatorPart):
		s.operatorRest()
	}
}

func (s *Scanner) operatorRest() {
	switch {
	case s.accept('/'):
		if !s.skipComment() {
			s.ingest('/')
		}
	case s.acceptIf(text.IsOperatorPart):
		s.operatorRest()
	case s.acceptIf(text.IsSpecial):
		s.operatorRest()
	}
}

////////////////////////////////////////////////////////////////////////////////
// Numbers

func (s *Scanner) base(digits fragments.Fragment, baseName string) (token text.Token) {
	if !s.atLeastOne(text.Digit) {
		s.syntaxError(s.currentPos(), "expected at least one digit in %s integer literal", baseName)
		token = s.wrapTokenAs(text.IntegerLit)
		return
	}
	content := s.text()
	if offset := strings.IndexFunc(content[2:], fragments.Not(digits)); offset >= 0 {
		pos := s.pos()
		pos.Column += offset
		s.syntaxError(pos, "unexpected digit in %s literal: '%c'", baseName, content[2:][offset])
	}
	token = s.wrapTokenAs(text.IntegerLit)
	return
}

func (s *Scanner) binary() (token text.Token) {
	return s.base(text.BinaryDigit, "binary")
}

func (s *Scanner) octal() (token text.Token) {
	return s.base(text.OctalDigit, "octal")
}

func (s *Scanner) decimal() (token text.Token) {
	return s.base(text.Digit, "decimal")
}

func (s *Scanner) hexadecimal() (token text.Token) {
	return s.base(text.HexDigit, "hexadecimal")
}

func (s *Scanner) exponent() (token text.Token) {
	if s.acceptIf(text.Exponent) {
		s.accept('+', '-') // optional
		if !s.atLeastOne(text.Digit) {
			s.syntaxError(s.currentPos(), "expected at least one exponent digit in float literal")
			content := s.data()
			cut := strings.LastIndexFunc(content, text.Exponent)
			token = s.wrapTokenWith(text.FloatLit, content[:cut])
			return
		}
	}
	token = s.wrapTokenAs(text.FloatLit)
	return
}

func (s *Scanner) decimalPart() (token text.Token) {
	if !s.atLeastOne(text.Digit) {
		s.syntaxError(s.currentPos(), "expected at least one digit after decimal point in float literal, found '%c'", s.peek())
	}
	token = s.exponent()
	return
}

func (s *Scanner) number() (token text.Token) {
	switch {
	case s.acceptIf(text.Digit):
		token = s.number()
	case s.accept('.'):
		token = s.decimalPart()
	case s.matchIf(text.Exponent):
		token = s.exponent()
	default:
		token = s.wrapTokenWith(text.IntegerLit, s.text())
	}
	return
}

////////////////////////////////////////////////////////////////////////////////
// String, Char & escapes

func (s *Scanner) escape(digits fragments.Fragment, expected int) {
	seq := text.EscapeBuffer(expected)
	for n := 0; n < expected; n++ {
		switch {
		case s.matchIf(digits):
			seq[n] = s.peek()
			s.skipRune()
		default:
			s.syntaxError(s.currentPos(), "invalid character in escape sequence (found %q, expected hexadecimal digit)", s.peek())
			return
		}
	}
	s.ingest(seq.Rune())
}

func (s *Scanner) invalidEscape() {
	s.syntaxError(s.currentPos(), "invalid escape character")
	s.advance()
}

func (s *Scanner) litRune() {
	switch {
	case s.match('\\'):
		s.skipRune()
		switch s.peek() {
		case 'b':
			s.ingest('\b')
		case 't':
			s.ingest('\t')
		case 'n':
			s.ingest('\n')
		case 'f':
			s.ingest('\f')
		case 'r':
			s.ingest('\r')
		case '"':
			s.ingest('"')
		case '\'':
			s.ingest('\'')
		case '\\':
			s.ingest('\\')
		case 'x', 'X':
			s.skipRune()
			s.escape(text.HexDigit, 2)
			return
		case 'u', 'U':
			s.skipRune()
			s.escape(text.HexDigit, 4)
			return
		default:
			s.invalidEscape()
		}
		s.skipRune()
	default:
		s.advance()
	}
}

func (s *Scanner) litRunes(del rune) {
	for !s.match(del) && !s.eof() && !s.match(text.SU, text.CR, text.LF) {
		s.litRune()
	}
}

func (s *Scanner) stringLit() {
	s.litRunes('"')
	if s.match('"') {
		s.skipRune()
	} else {
		s.syntaxError(s.currentPos(), "unclosed stringLit literal")
	}
}

func (s *Scanner) charLit() {
	s.skipRune()
	switch {
	case s.matchIf(text.IsIdentifierStart):
		s.charLitOr(s.identRest)
	case s.matchIf(text.IsOperatorPart) && !s.match('\\'):
		s.charLitOr(s.operatorRest)
	case !s.eof() && !s.match(text.SU, text.CR, text.LF):
		emptyCharLit := s.match('\'')
		s.litRune()
		switch {
		case s.match('\''):
			if emptyCharLit {
				s.syntaxError(s.pos(), "empty character literal (use '\\'' for single quote)")
			} else {
				s.skipRune()
			}
		case emptyCharLit:
			s.syntaxError(s.pos(), "empty character literal")
		default:
			s.syntaxError(s.currentPos(), "unclosed character literal")
		}
	default:
		s.syntaxError(s.currentPos(), "unclosed character literal")
	}
}

func (s *Scanner) charLitOr(op func()) {

}
