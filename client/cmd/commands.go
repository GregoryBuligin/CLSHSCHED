package cmd

import (
	"github.com/urfave/cli"
)

var generalFlags = []cli.Flag{
	&cli.IntFlag{
		Name:   "port",
		Usage:  "port to bind to",
		EnvVar: "TURNIP_PORT",
		Value:  8000,
	},
	&cli.StringFlag{
		Name:   "host",
		Usage:  "port to bind to",
		EnvVar: "TURNIP_HOST",
		Value:  "turnip.drw",
	},
}

var SchedTaskCommand = cli.Command{
	Name:      "sched",
	Usage:     "sched task between cluster nodes",
	UsageText: "./clctl sched --recipe=Recipe.json",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "recipe, r, f",
			Usage: "Recipe.json (required)",
			Value: "",
		},
	},
	Action: SchedTaskCommandAction,
}
