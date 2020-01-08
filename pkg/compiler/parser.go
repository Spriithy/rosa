package compiler

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Spriithy/rosa/pkg/compiler/text"
)

type Parser struct {
	path    string
	Scanner *Scanner
	tokens  *[]text.Token
	current int
	Logs    []Log
}

func NewParser(path string) *Parser {
	p := &Parser{
		path:    path,
		Scanner: NewScanner(path),
	}
	for token := p.Scanner.Scan(); !text.Eof(token); token = p.Scanner.Scan() {
		fmt.Println(token.String())
	}
	p.tokens = &p.Scanner.tokens
	return p
}

func (p *Parser) error(token text.Token, err error) {
	p.Logs = append(p.Logs, Log{
		Level:   LogError,
		Message: err.Error(),
		Pos:     token.Pos,
	})
}

func (p *Parser) errorf(token text.Token, message string, args ...interface{}) {
	p.Logs = append(p.Logs, Log{
		Level:   LogError,
		Message: fmt.Sprintf(message, args...),
		Pos:     token.Pos,
	})
}

////////////////////////////////////////////////////////////////////////////////

func (p *Parser) eof() bool {
	return p.current >= len(*p.tokens)
}

func (p *Parser) peek(n int) text.Token {
	if p.current+n >= len(*p.tokens) {
		return p.previous()
	}
	return (*p.tokens)[p.current+n]
}

func (p *Parser) previous() text.Token {
	return p.peek(-1)
}

func (p *Parser) lookahead() text.Token {
	if p.eof() {
		return p.previous()
	}
	return p.peek(0)
}

func (p *Parser) advance() text.Token {
	if !p.eof() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) match(fs ...func(text.Token) bool) bool {
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

func (p *Parser) expect(fs ...func(text.Token) bool) func(string) (text.Token, error) {
	return func(message string) (token text.Token, err error) {
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
		if text.Semicolon(p.previous()) {
			return
		}
		switch token := p.lookahead(); {
		case text.Module(token):
			return
		case text.Import(token):
			return
		case text.Def(token):
			return
		case text.Struct(token):
			return
		case text.Trait(token):
			return
		case text.Let(token):
			return
		case text.Match(token):
			return
		case text.Case(token):
			return
		case text.Return(token):
			return
		default:
			p.advance()
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

func (p *Parser) Parse() AST {
	return p.compilationUnit()
}

func (p *Parser) compilationUnit() (module *ModuleAST) {
	if !p.match(text.Module) {
		p.errorf(p.peek(0), "expected module declaration")
		return &ModuleAST{
			Name: "<invalid>",
		}
	}
	moduleToken := p.previous()
	moduleName, err := p.expect(text.Identifier)("expected module name after 'module' token")
	if err != nil {
		p.error(moduleToken, err)
	}
	module = &ModuleAST{
		Tokens: []text.Token{moduleToken, moduleName},
		Name:   moduleName.Text,
	}
	for p.def(module) != nil {
	}
	return
}

func (p *Parser) def(module *ModuleAST) (decl *DeclAST) {
	switch {
	case p.match(text.Def):
		decl = &DeclAST{}
		defName, err := p.expect(text.Identifier)("expected identifier")
		if err != nil {
			p.error(defName, err)
			return
		}
		decl.Name = defName.Text
		if p.match(text.Assign) {
			decl.Expr = p.literal()
		} else {
			fmt.Println(p.peek(0))
		}
		module.Decls = append(module.Decls, decl)
	}
	return
}

func (p *Parser) literal() (expr Expr) {
	switch {
	case p.match(text.Minus):
		switch {
		case p.match(text.Integer):
			token := p.previous()
			value, _ := strconv.ParseInt(token.Text, 0, 64)
			expr = &SignedIntegerExpr{
				Token: token,
				Value: -value,
			}
		case p.match(text.Float):
			token := p.previous()
			value, _ := strconv.ParseFloat(token.Text, 64)
			expr = &FloatExpr{
				Token: token,
				Value: -value,
			}
		default:
			p.errorf(p.lookahead(), "expected integer or float literal, instead found '%s'", p.lookahead().Text)
		}
		return
	case p.match(text.Boolean):
		token := p.previous()
		value, _ := strconv.ParseBool(token.Text)
		expr = &BooleanExpr{
			Token: token,
			Value: value,
		}
		return
	case p.match(text.Integer):
		token := p.previous()
		value, _ := strconv.ParseUint(token.Text, 0, 64)
		expr = &UnsignedIntegerExpr{
			Token: token,
			Value: value,
		}
		return
	case p.match(text.Float):
		token := p.previous()
		value, _ := strconv.ParseFloat(token.Text, 64)
		expr = &FloatExpr{
			Token: token,
			Value: value,
		}
		return
	case p.match(text.String):
		expr = &StringExpr{
			Token: p.previous(),
			Value: p.previous().Text,
		}
		return
	}
	p.errorf(p.lookahead(), "expected expression, found '%s'", p.lookahead().Text)
	return
}
