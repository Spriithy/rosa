package compiler

import "errors"

type Log struct {
	Level   string
	Message string
	Pos     Pos
}

const (
	LogError   = "error"
	LogWarning = "warning"
)

func (l Log) asError() error {
	return errors.New(l.Message)
}
