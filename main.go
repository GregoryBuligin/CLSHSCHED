package main

import (
	"fmt"
	"os"
	cmdserver "shsched/shsched/cmd"

	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:                 "clshshed",
		Usage:                "clshshed serve --port=8001",
		EnableBashCompletion: true,
		Commands: []cli.Command{
			cmdserver.ServeCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}
}
