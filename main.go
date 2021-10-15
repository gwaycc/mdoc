package main

import (
	"os"

	"github.com/gwaycc/mdoc/cmd"

	"github.com/urfave/cli/v2"
)

var app = &cmd.App{
	&cli.App{
		Name:    "Markdown Document",
		Version: cmd.Version(),
		Usage:   "Run mdoc server",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "repo",
				Value: "./data",
				Usage: "data repo",
			},
		},
	},
}

func main() {
	if err := app.Run(os.Args); err != nil {
		panic(err)
	}
}
