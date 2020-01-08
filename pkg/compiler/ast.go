package compiler

import "github.com/Spriithy/rosa/pkg/compiler/text"

type AST interface {
	Accept(AstPrinter) string
	ast()
}

////////////////////////////////////////////////////////////////////////////////

type ModuleAST struct {
	Name   string
	Tokens []text.Token
	Decls  []*DeclAST
}

func (*ModuleAST) ast()                               {}
func (m *ModuleAST) Accept(printer AstPrinter) string { return printer.visitModuleAST(m) }

////////////////////////////////////////////////////////////////////////////////

type DeclAST struct {
	Name   string
	Tokens []text.Token
	Expr   Expr
}

func (*DeclAST) ast()                               {}
func (d *DeclAST) Accept(printer AstPrinter) string { return printer.visitDeclAST(d) }
