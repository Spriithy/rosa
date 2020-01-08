package compiler

import (
	"errors"

	"github.com/Spriithy/rosa/pkg/compiler/text"
)

type Log struct {
	Level   string
	Message string
	Pos     text.Pos
}

const (
	LogError   = "error"
	LogWarning = "warning"
)

func (l Log) asError() error {
	return errors.New(l.Message)
}
