package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/urfave/cli"
	"google.golang.org/grpc/status"

	"shsched/netscanner"
	"shsched/shsched"
)

// GetNewClientGRPC is shortcut function for creation GRPC-client
func GetNewClientGRPC(
	ctx context.Context,
	port int,
) (context.Context, *shsched.Client, error) {
	address, err := netscanner.ExternalIP()
	if err != nil {
		return ctx, nil, err
	}

	myHost, _, err := netscanner.ScanMyIP(ctx, address)
	if err != nil {
		return ctx, nil, err
	}

	client, err := shsched.NewClient(&shsched.ClientConfig{
		Address:   fmt.Sprintf("%s:%d", myHost, port),
		UseLogger: false,
	})
	if err != nil {
		return ctx, nil, err
	}

	return ctx, client, nil
}

// SchedTaskCommandAction starts "sched" client command
func SchedTaskCommandAction(c *cli.Context) (err error) {
	if c.NumFlags() < 1 {
		if err = cli.ShowCommandHelp(c, "sched"); err != nil {
			return err
		}
	}

	// Get GRPC client
	ctx, client, err := GetNewClientGRPC(context.Background(), c.Int("port"))
	if err != nil {
		return err
	}
	defer client.Close()

	recipePath := c.String("recipe")
	if recipePath == "" {
		return errors.New("recipe location not set")
	}

	// GRPC-call
	_, err = client.SchedTask(ctx, recipePath)
	if err != nil {
		statusCode, ok := status.FromError(err)
		if ok {
			return errors.New(statusCode.Message())
		}

		return err
	}

	fmt.Println(">>>", "Success sched")

	return nil
}
