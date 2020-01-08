package compiler

import "fmt"

type Token struct {
	Text string
	Type string
	Pos
}

func (t Token) String() string {
	return fmt.Sprintf("Token{(%d,%d), %s, '%s'}", t.Line, t.Col, t.Type, t.Text)
}

const (
	Error      = "ERROR"
	EOF        = "EOF"
	Identifier = "IDENTIFIER"
	Keyword    = "KEYWORD"
	Integer    = "INTEGER"
	Float      = "FLOAT"
	String     = "STRING"
	Separator  = "SEPARATOR"
	Operator   = "OPERATOR"
)

func tokenOf(typ string) func(Token) bool {
	return func(tok Token) bool { return tok.Type == typ }
}

func typedText(typ string, text string) func(Token) bool {
	return func(tok Token) bool { return tokenOf(typ)(tok) && tok.Text == text }
}

func sep(sep string) func(Token) bool {
	separators[sep] = true
	return typedText(Separator, sep)
}

func op(op string) func(Token) bool {
	operators[op] = true
	return typedText(Operator, op)
}

func keyword(keyword string) func(Token) bool {
	keywords[keyword] = true
	return typedText(Keyword, keyword)
}

func anyOf(fs ...func(Token) bool) func(Token) bool {
	return func(tok Token) bool {
		for _, f := range fs {
			if f(tok) {
				return true
			}
		}
		return false
	}
}

var (
	keywords   = map[string]bool{}
	separators = map[string]bool{}
	operators  = map[string]bool{}
)

func tokenType(str string) (typ string) {
	switch {
	case keywords[str]:
		typ = Keyword
	case separators[str]:
		typ = Separator
	case operators[str]:
		typ = Operator
	default:
		typ = Identifier
	}
	return
}

var (
	eof     = tokenOf(EOF)
	ident   = tokenOf(Identifier)
	integer = tokenOf(Integer)
	float   = tokenOf(Float)
	str     = tokenOf(String)

	module       = keyword("module")
	importRule   = keyword("import")
	trait        = keyword("trait")
	strukt       = keyword("struct")
	typ          = keyword("type")
	def          = keyword("def")
	let          = keyword("let")
	declKeyword  = anyOf(typ, def, let)
	mut          = keyword("mut")
	ret          = keyword("return")
	match        = keyword("match")
	matchCase    = keyword("case")
	boolTrue     = keyword("true")
	boolFalse    = keyword("false")
	boolean      = anyOf(boolTrue, boolFalse)
	literalValue = anyOf(integer, float, str, boolean)

	comma     = sep(",")
	colon     = sep(":")
	semicolon = sep(";")
	lpar      = sep("(")
	rpar      = sep(")")
	lbrk      = sep("[")
	rbrk      = sep("]")
	lbrc      = sep("{")
	rbrc      = sep("}")

	arrow    = op("=>")
	dot      = op(".")
	plus     = op("+")
	inc      = op("++")
	minus    = op("-")
	dec      = op("--")
	star     = op("*")
	div      = op("/")
	and      = op("&")
	land     = op("&&")
	or       = op("|")
	lor      = op("||")
	xor      = op("^")
	not      = op("~")
	lnot     = op("!")
	lt       = op("<")
	lte      = op("<=")
	gt       = op(">")
	gte      = op(">=")
	eq       = op("==")
	neq      = op("!=")
	assign   = op("=")
	walrus   = op(":=")
	binaryOp = anyOf(plus, minus, star, div, and, or, land, lor, xor, lt, lte, gt, gte)
	unaryOp  = anyOf(minus, inc, dec, not, lnot)

	logicalBinaryOp = anyOf(eq, land, lor, xor)
	logicalUnaryOp  = lnot
)
