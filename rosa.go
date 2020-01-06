package main

import (
	"fmt"
	"github.com/Spriithy/rosa/commands"
	"github.com/urfave/cli"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "rosa"
	app.Usage = "The rosa programming language"
	app.Commands = []*cli.Command{
		commands.BuildCommand(),
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
