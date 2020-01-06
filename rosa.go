package main

import (
	"fmt"

	"github.com/Spriithy/rosa/pkg/compiler"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rosa"
	app.Usage = "The rosa programming language"

	s := compiler.NewScanner("examples/scanner.rosa")
	token := s.Scan()
	for token.Type != compiler.EOF {
		fmt.Println(token.String())
		token = s.Scan()
	}
	fmt.Println(token.String())
}
