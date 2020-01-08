package ast

import (
	"fmt"
	"strings"

	"github.com/Spriithy/rosa/pkg/compiler/text"
)

type Expr interface {
	expr()
	AST
}

////////////////////////////////////////////////////////////////////////////////

type AstPrinter struct{}

func (p AstPrinter) parenthesize(name string, asts ...AST) string {
	var sb strings.Builder
	sb.WriteByte('(')
	sb.WriteString(name)
	for _, ast := range asts {
		sb.WriteByte(' ')
		sb.WriteString(ast.Accept(p))
	}
	sb.WriteRune(')')
	return sb.String()
}

func (p AstPrinter) visitModuleAST(ast *ModuleAST) string {
	asts := make([]AST, len(ast.Decls))
	for i := range ast.Decls {
		asts[i] = ast.Decls[i]
	}
	return p.parenthesize("module "+ast.Name+"\n", asts...)
}

func (p AstPrinter) visitDeclAST(ast *DeclAST) string {
	return p.parenthesize("decl "+ast.Name, ast.Expr) + "\n"
}

func (p AstPrinter) visitBinaryExpr(expr *BinaryExpr) string {
	return p.parenthesize(expr.Op.Text, expr.Left, expr.Right)
}

func (p AstPrinter) visitUnaryExpr(expr *UnaryExpr) string {
	return p.parenthesize(expr.Op.Text, expr.Expr)
}

func (p AstPrinter) visitGroupingExpr(expr *GroupingExpr) string {
	return p.parenthesize("group", expr.Expr)
}

func (p AstPrinter) visitBooleanExpr(expr *BooleanExpr) string {
	return expr.Token.Text
}

func (p AstPrinter) visitSignedIntegerExpr(expr *SignedIntegerExpr) string {
	return fmt.Sprintf("%d", expr.Value)
}

func (p AstPrinter) visitUnsignedIntegerExpr(expr *UnsignedIntegerExpr) string {
	return fmt.Sprintf("%d", expr.Value)
}

func (p AstPrinter) visitFloatExpr(expr *FloatExpr) string {
	return fmt.Sprintf("%f", expr.Value)
}

func (p AstPrinter) visitStringExpr(expr *StringExpr) string {
	return fmt.Sprintf("\"%s\"", expr.Value)
}

func (p AstPrinter) visitIdentExpr(expr *IdentExpr) string {
	return expr.Name
}

////////////////////////////////////////////////////////////////////////////////

type BinaryExpr struct {
	Left  Expr
	Op    text.Token
	Right Expr
}

func (*BinaryExpr) ast()                            {}
func (*BinaryExpr) expr()                           {}
func (expr *BinaryExpr) Accept(p AstPrinter) string { return p.visitBinaryExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type UnaryExpr struct {
	Op   text.Token
	Expr Expr
}

func (*UnaryExpr) ast()                            {}
func (*UnaryExpr) expr()                           {}
func (expr *UnaryExpr) Accept(p AstPrinter) string { return p.visitUnaryExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type GroupingExpr struct {
	Lpar text.Token
	Expr Expr
	Rpar text.Token
}

func (*GroupingExpr) ast()                            {}
func (*GroupingExpr) expr()                           {}
func (expr *GroupingExpr) Accept(p AstPrinter) string { return p.visitGroupingExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type BooleanExpr struct {
	Token text.Token
	Value bool
}

func (*BooleanExpr) ast()                            {}
func (*BooleanExpr) expr()                           {}
func (expr *BooleanExpr) Accept(p AstPrinter) string { return p.visitBooleanExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type SignedIntegerExpr struct {
	Token text.Token
	Value int64
}

func (*SignedIntegerExpr) ast()                            {}
func (*SignedIntegerExpr) expr()                           {}
func (expr *SignedIntegerExpr) Accept(p AstPrinter) string { return p.visitSignedIntegerExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type UnsignedIntegerExpr struct {
	Token text.Token
	Value uint64
}

func (*UnsignedIntegerExpr) ast()                            {}
func (*UnsignedIntegerExpr) expr()                           {}
func (expr *UnsignedIntegerExpr) Accept(p AstPrinter) string { return p.visitUnsignedIntegerExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type FloatExpr struct {
	Token text.Token
	Value float64
}

func (*FloatExpr) ast()                            {}
func (*FloatExpr) expr()                           {}
func (expr *FloatExpr) Accept(p AstPrinter) string { return p.visitFloatExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type StringExpr struct {
	Token text.Token
	Value string
}

func (*StringExpr) ast()                            {}
func (*StringExpr) expr()                           {}
func (expr *StringExpr) Accept(p AstPrinter) string { return p.visitStringExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type IdentExpr struct {
	Token text.Token
	Name  string
}

func (*IdentExpr) ast()                            {}
func (*IdentExpr) expr()                           {}
func (expr *IdentExpr) Accept(p AstPrinter) string { return p.visitIdentExpr(expr) }

////////////////////////////////////////////////////////////////////////////////
