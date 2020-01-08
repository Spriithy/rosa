package compiler

import "strings"

import "fmt"

////////////////////////////////////////////////////////////////////////////////

type AST interface {
	Accept(AstPrinter) string
	ast()
}

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

func (p AstPrinter) visitIntegerExpr(expr *IntegerExpr) string {
	return expr.Token.Text
}

func (p AstPrinter) visitFloatExpr(expr *FloatExpr) string {
	return expr.Token.Text
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
	Op    Token
	Right Expr
}

func (*BinaryExpr) ast()                            {}
func (*BinaryExpr) expr()                           {}
func (expr *BinaryExpr) Accept(p AstPrinter) string { return p.visitBinaryExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type UnaryExpr struct {
	Op   Token
	Expr Expr
}

func (*UnaryExpr) ast()                            {}
func (*UnaryExpr) expr()                           {}
func (expr *UnaryExpr) Accept(p AstPrinter) string { return p.visitUnaryExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type GroupingExpr struct {
	Lpar Token
	Expr Expr
	Rpar Token
}

func (*GroupingExpr) ast()                            {}
func (*GroupingExpr) expr()                           {}
func (expr *GroupingExpr) Accept(p AstPrinter) string { return p.visitGroupingExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type BooleanExpr struct {
	Token Token
	Value bool
}

func (*BooleanExpr) ast()                            {}
func (*BooleanExpr) expr()                           {}
func (expr *BooleanExpr) Accept(p AstPrinter) string { return p.visitBooleanExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type IntegerExpr struct {
	Token Token
	Value uint64
}

func (*IntegerExpr) ast()                            {}
func (*IntegerExpr) expr()                           {}
func (expr *IntegerExpr) Accept(p AstPrinter) string { return p.visitIntegerExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type FloatExpr struct {
	Token Token
	Value float64
}

func (*FloatExpr) ast()                            {}
func (*FloatExpr) expr()                           {}
func (expr *FloatExpr) Accept(p AstPrinter) string { return p.visitFloatExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type StringExpr struct {
	Token Token
	Value string
}

func (*StringExpr) ast()                            {}
func (*StringExpr) expr()                           {}
func (expr *StringExpr) Accept(p AstPrinter) string { return p.visitStringExpr(expr) }

////////////////////////////////////////////////////////////////////////////////

type IdentExpr struct {
	Token Token
	Name  string
}

func (*IdentExpr) ast()                            {}
func (*IdentExpr) expr()                           {}
func (expr *IdentExpr) Accept(p AstPrinter) string { return p.visitIdentExpr(expr) }

////////////////////////////////////////////////////////////////////////////////
