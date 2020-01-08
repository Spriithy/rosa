package compiler

import "strconv"

import "errors"

import "fmt"

type Parser struct {
	path    string
	Scanner *Scanner
	tokens  *[]Token
	current int
	Logs    []Log
}

func NewParser(path string) *Parser {
	p := &Parser{
		path:    path,
		Scanner: NewScanner(path),
	}
	for token := p.Scanner.Scan(); !eof(token); token = p.Scanner.Scan() {
		fmt.Println(token.String())
	}
	p.tokens = &p.Scanner.tokens
	return p
}

func (p *Parser) error(token Token, err error) {
	p.Logs = append(p.Logs, Log{
		Level:   LogError,
		Message: err.Error(),
		Pos:     token.Pos,
	})
}

func (p *Parser) errorf(token Token, message string, args ...interface{}) {
	p.Logs = append(p.Logs, Log{
		Level:   LogError,
		Message: fmt.Sprintf(message, args...),
		Pos:     token.Pos,
	})
}

////////////////////////////////////////////////////////////////////////////////

func (p *Parser) eof() bool {
	return eof((*p.tokens)[len(*p.tokens)-1])
}

func (p *Parser) peek(n int) Token {
	if p.current+n >= len(*p.tokens) {
		return p.previous()
	}
	return (*p.tokens)[p.current+n]
}

func (p *Parser) previous() Token {
	return p.peek(-1)
}

func (p *Parser) lookahead() Token {
	if p.eof() {
		return p.previous()
	}
	return p.peek(0)
}

func (p *Parser) advance() Token {
	if !p.eof() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) match(fs ...func(Token) bool) bool {
	if p.eof() {
		return false
	}
	for _, f := range fs {
		if f(p.lookahead()) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) expect(fs ...func(Token) bool) func(string) (Token, error) {
	return func(message string) (token Token, err error) {
		for _, f := range fs {
			if f(p.lookahead()) {
				p.advance()
				token = p.previous()
				return
			}
		}
		token = p.lookahead()
		err = errors.New(message)
		return
	}
}

////////////////////////////////////////////////////////////////////////////////

func (p *Parser) sync() {
	p.advance()
	for !p.eof() {
		if semicolon(p.previous()) {
			return
		}
		switch token := p.lookahead(); {
		case module(token):
			return
		case importRule(token):
			return
		case def(token):
			return
		case strukt(token):
			return
		case trait(token):
			return
		case let(token):
			return
		case match(token):
			return
		case matchCase(token):
			return
		case ret(token):
			return
		default:
			p.advance()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (p *Parser) Parse() AST {
	return p.expr()
}

func (p *Parser) expr() Expr {
	return p.equality()
}

func (p *Parser) equality() Expr {
	expr := p.comparison()
	for p.match(eq, neq) {
		op := p.previous()
		right := p.comparison()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}
	return expr
}

func (p *Parser) comparison() Expr {
	expr := p.additive()
	for p.match(gt, gte, lt, lte) {
		op := p.previous()
		right := p.additive()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}
	return expr
}

func (p *Parser) additive() Expr {
	expr := p.multiplicative()
	for p.match(plus, minus) {
		op := p.previous()
		right := p.multiplicative()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}
	return expr
}

func (p *Parser) multiplicative() Expr {
	expr := p.unary()
	for p.match(star, div) {
		op := p.previous()
		right := p.unary()
		expr = &BinaryExpr{
			Left:  expr,
			Op:    op,
			Right: right,
		}
	}
	return expr
}

func (p *Parser) unary() Expr {
	if p.match(not, lnot, minus) {
		op := p.previous()
		expr := p.unary()
		return &UnaryExpr{
			Op:   op,
			Expr: expr,
		}
	}
	return p.literal()
}

func (p *Parser) literal() (expr Expr) {
	switch {
	case p.match(boolean):
		token := p.previous()
		value, _ := strconv.ParseBool(token.Text)
		expr = &BooleanExpr{
			Token: token,
			Value: value,
		}
		return
	case p.match(integer):
		token := p.previous()
		value, _ := strconv.ParseUint(token.Text, 0, 64)
		expr = &IntegerExpr{
			Token: token,
			Value: value,
		}
		return
	case p.match(float):
		token := p.previous()
		value, _ := strconv.ParseFloat(token.Text, 64)
		expr = &FloatExpr{
			Token: token,
			Value: value,
		}
		return
	case p.match(str):
		expr = &StringExpr{
			Token: p.previous(),
			Value: p.previous().Text,
		}
		return
	case p.match(lpar):
		lpar := p.previous()
		expr = p.expr()
		rpar, err := p.expect(rpar)("expected closing parenthesis ')' after expression")
		if err != nil {
			p.error(rpar, err)
		}
		expr = &GroupingExpr{
			Lpar: lpar,
			Expr: expr,
			Rpar: rpar,
		}
		return
	}
	p.errorf(p.lookahead(), "expected expression, found '%s'", p.lookahead().Text)
	return
}
