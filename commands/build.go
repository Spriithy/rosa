package commands

import (
	"errors"
	"fmt"

	"github.com/Spriithy/rosa/pkg/compiler"
	"github.com/Spriithy/rosa/pkg/compiler/text"
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

func buildAction(c *cli.Context) (err error) {
	if !c.Args().Present() {
		err = NoInputFileError
		return
	}

	file := c.Args().First()
	/*p := compiler.NewParser(file)
	result := p.Parse()
	tree := result.Accept(ast.AstPrinter{})
	fmt.Println(tree)
	*/
	s := compiler.NewScanner(file)
	for token := s.Scan(); !text.Eof(token); token = s.Scan() {
		fmt.Println(token.String())
	}
	for _, log := range s.Logs {
		fmt.Println(log.AsError())
	}
	return
}
