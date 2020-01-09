package compiler

import (
	"fmt"

	"github.com/Spriithy/rosa/pkg/compiler/text"
)

type Log struct {
	Path    string
	Level   string
	Message string
	Pos     text.Pos
}

const (
	LogSyntaxError = "syntax error"
	LogError       = "error"
)

func (l Log) AsError() error {
	return fmt.Errorf("%s:%d:%d: %s: %s", l.Path, l.Pos.Line, l.Pos.Col, l.Level, l.Message)
}
