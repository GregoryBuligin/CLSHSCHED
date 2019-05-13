package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"

	"shsched/client/cmd"
)

func main() {
	app := &cli.App{
		Name:                 "clctl",
		Usage:                "clctl -h",
		EnableBashCompletion: true,
		Commands: []cli.Command{
			cmd.SchedTaskCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		os.Exit(1)
	}
}
