package text

import (
	"fmt"
)

type Pos struct {
	FileName string
	Line     int
	Column   int
}

func (pos Pos) String() (s string) {
	if pos.FileName != "" {
		s += pos.FileName + ":"
	}
	if pos.Line > 0 {
		s += fmt.Sprintf("%d", pos.Line)
		if pos.Column != 0 {
			s += fmt.Sprintf(":%d", pos.Column)
		}
	}
	return
}

type TokenType struct {
	Name       string
	Paraphrase string
}

func (t TokenType) IsValid() bool {
	return t.Name != ""
}

func (t TokenType) Equals(other TokenType) bool {
	return t.Name == other.Name
}

func (t TokenType) String() string {
	if t.Paraphrase != "" {
		return t.Paraphrase
	}
	return t.Name
}

type Token struct {
	Type  TokenType
	Text  string
	Spans int
	Pos
}

func (t Token) String() string {
	return fmt.Sprintf("%s: %s: %q", t.Pos, t.Type.Name, t.Text)
}

var (
	tokenTypes = map[string]TokenType{}

	// Generic Token types

	ErrorType      = registerTokenType("Error", "an error")
	EOF            = registerTokenType("Eof", "end of file")
	IdentifierType = registerTokenType("Identifier", "an identifier")
	KeywordType    = registerTokenType("Keyword", "a keyword")
	IntegerLit     = registerTokenType("IntegerLit", "an integer")
	FloatLit       = registerTokenType("FloatLit", "a float")
	CharLit        = registerTokenType("CharLit", "a character literal")
	StringLit      = registerTokenType("StringLit", "a string literal")
	SeparatorType  = registerTokenType("Separator", "a separator")
	OperatorType   = registerTokenType("Operator", "an operator")

	// Specific Token types

	LparType      = registerCharToken('(')
	RparType      = registerCharToken(')')
	LbrkType      = registerCharToken('[')
	RbrkType      = registerCharToken(']')
	LbrcType      = registerCharToken('{')
	RbrcType      = registerCharToken('}')
	ColonType     = registerCharToken(':')
	SemicolonType = registerCharToken(';')
	CommaType     = registerCharToken(',')
	DotType       = registerCharToken('.')
)

func registerCharToken(r rune) TokenType {
	name := fmt.Sprintf("%c", r)
	paraphrase := fmt.Sprintf("%q", r)
	return registerTokenType(name, paraphrase)
}

func registerTokenType(name, paraphrase string) (typ TokenType) {
	typ = TokenType{
		Name:       name,
		Paraphrase: paraphrase,
	}
	tokenTypes[name] = typ
	return
}

func tokenOf(typ TokenType) func(Token) bool {
	return func(tok Token) bool { return tok.Type == typ }
}

func typedText(typ TokenType, text string) func(Token) bool {
	return func(tok Token) bool { return tokenOf(typ)(tok) && tok.Text == text }
}

func sep(sep string) func(Token) bool {
	separators[sep] = true
	if typ := tokenTypes[sep]; typ.IsValid() {
		return typedText(typ, sep)
	}
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

func TypeOfToken(str string) (typ TokenType) {
	switch {
	case tokenTypes[str].IsValid():
		typ = tokenTypes[str]
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
	Eof        = tokenOf(EOF)
	Identifier = tokenOf(IdentifierType)
	Integer    = tokenOf(IntegerLit)
	Float      = tokenOf(FloatLit)
	Char       = tokenOf(CharLit)
	String     = tokenOf(StringLit)

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
	Literal = anyOf(Integer, Float, String, Char, Boolean)

	Comma     = tokenOf(CommaType)
	Colon     = tokenOf(ColonType)
	Semicolon = tokenOf(SemicolonType)
	Lpar      = tokenOf(LparType)
	Rpar      = tokenOf(RparType)
	Lbrk      = tokenOf(LbrkType)
	Rbrk      = tokenOf(RbrkType)
	Lbrc      = tokenOf(LbrcType)
	Rbrc      = tokenOf(RbrcType)

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
