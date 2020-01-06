package commands

import (
	"errors"
	"fmt"

	"github.com/Spriithy/rosa/pkg/compiler"
	"github.com/urfave/cli"
)

var (
	NoInputFileError = errors.New("no input file")
)

func BuildCommand() *cli.Command {
	return &cli.Command{
		Name:   "build",
		Usage:  "Build a rosa module",
		Action: buildAction,
	}
}

func buildAction(c *cli.Context) error {
	if !c.Args().Present() {
		return NoInputFileError
	}

	file := c.Args().First()
	s := compiler.NewScanner(file)

	token := s.Scan()
	for token.Type != compiler.EOF {
		fmt.Println(token.String())
		token = s.Scan()
	}
	fmt.Println(token.String())

	for _, log := range s.Logs {
		fmt.Printf("%s: %s\n", log.Level, log.Message)
	}

	return nil
}
