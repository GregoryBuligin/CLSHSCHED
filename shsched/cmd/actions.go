package cmdserver

import (
	"errors"
	"strconv"

	"github.com/urfave/cli"

	"shsched/shsched"
)

func ServeCommandAction(c *cli.Context) (err error) {
	if c.NumFlags() < 1 {
		if err = cli.ShowCommandHelp(c, "serve"); err != nil {
			return err
		}
	}

	port := c.Int("port")
	if port == 0 {
		return errors.New("port is not set")
	}

	cfg := &shsched.ServerConfig{
		Port:      strconv.Itoa(port),
		UseLogger: c.Bool("debug"),
	}

	server, err := shsched.NewServer(cfg)
	if err != nil {
		return err
	}

	var semaphore = make(chan uint, 100)

	go server.OutputWaiter()
	go server.SelectTask()
	go shsched.StartRunner(server, semaphore)

	if err := server.Serve(); err != nil {
		return err
	}

	return nil
}
