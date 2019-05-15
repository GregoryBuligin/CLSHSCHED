package cmdserver

import (
	"github.com/urfave/cli"
)

var generalFlags = []cli.Flag{
	&cli.IntFlag{
		Name:   "port",
		Usage:  "port to bind to",
		EnvVar: "CLSHSHED_PORT",
	},
	&cli.BoolFlag{
		Name:   "debug",
		Usage:  "--debug",
		Hidden: true,
	},
}

var ServeCommand = cli.Command{
	Name:      "serve",
	Usage:     "sched task between cluster nodes",
	UsageText: "./clshshed serve --port=8001",
	Flags:     generalFlags,
	Action:    ServeCommandAction,
}
