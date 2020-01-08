package text

import "fmt"

type Pos struct {
	Line int
	Col  int
}

type Token struct {
	Text string
	Type string
	Pos
}

func (t Token) String() string {
	return fmt.Sprintf("Token{(%d,%d), %s, '%s'}", t.Line, t.Col, t.Type, t.Text)
}

const (
	ErrorType      = "ERROR"
	EofType        = "EOF"
	IdentifierType = "IDENTIFIER"
	KeywordType    = "KEYWORD"
	IntegerType    = "INTEGER"
	FloatType      = "FLOAT"
	StringType     = "STRING"
	SeparatorType  = "SEPARATOR"
	OperatorType   = "OPERATOR"
)

func tokenOf(typ string) func(Token) bool {
	return func(tok Token) bool { return tok.Type == typ }
}

func typedText(typ string, text string) func(Token) bool {
	return func(tok Token) bool { return tokenOf(typ)(tok) && tok.Text == text }
}

func sep(sep string) func(Token) bool {
	separators[sep] = true
	return typedText(SeparatorType, sep)
}

func op(op string) func(Token) bool {
	operators[op] = true
	return typedText(OperatorType, op)
}

func keyword(keyword string) func(Token) bool {
	keywords[keyword] = true
	return typedText(KeywordType, keyword)
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

func TokenType(str string) (typ string) {
	switch {
	case keywords[str]:
		typ = KeywordType
	case separators[str]:
		typ = SeparatorType
	case operators[str]:
		typ = OperatorType
	default:
		typ = IdentifierType
	}
	return
}

var (
	Eof        = tokenOf(EofType)
	Identifier = tokenOf(IdentifierType)
	Integer    = tokenOf(IntegerType)
	Float      = tokenOf(FloatType)
	String     = tokenOf(StringType)

	Module  = keyword("module")
	Import  = keyword("import")
	Trait   = keyword("trait")
	Struct  = keyword("struct")
	Type    = keyword("type")
	Def     = keyword("def")
	Let     = keyword("let")
	Decl    = anyOf(Type, Def, Let)
	Mut     = keyword("mut")
	Return  = keyword("return")
	Match   = keyword("match")
	Case    = keyword("case")
	True    = keyword("true")
	False   = keyword("false")
	Boolean = anyOf(True, False)
	Literal = anyOf(Integer, Float, String, Boolean)

	Comma     = sep(",")
	Colon     = sep(":")
	Semicolon = sep(";")
	Lpar      = sep("(")
	Rpar      = sep(")")
	Lbrk      = sep("[")
	Rbrk      = sep("]")
	Lbrc      = sep("{")
	Rbrc      = sep("}")

	Arrow    = op("=>")
	Dot      = op(".")
	Plus     = op("+")
	Inc      = op("++")
	Minus    = op("-")
	Dec      = op("--")
	Star     = op("*")
	Div      = op("/")
	And      = op("&")
	Land     = op("&&")
	Or       = op("|")
	Lor      = op("||")
	Xor      = op("^")
	Not      = op("~")
	Lnot     = op("!")
	Lt       = op("<")
	Lte      = op("<=")
	Gt       = op(">")
	Gte      = op(">=")
	Eq       = op("==")
	Neq      = op("!=")
	Assign   = op("=")
	Walrus   = op(":=")
	BinaryOp = anyOf(Plus, Minus, Star, Div, And, Or, Land, Lor, Xor, Lt, Lte, Gt, Gte)
	UnaryOp  = anyOf(Minus, Inc, Dec, Not, Lnot)
)
