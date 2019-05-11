package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"google.golang.org/grpc"

	"github.com/rs/zerolog"

	types "shsched/shsched"
	shsched "shsched/shsched/server"
)

const (
	address = "localhost:8000"
)

type ClientConfig struct {
	Address string
}

type Client struct {
	address    string
	client     shsched.ShschedClient
	connection *grpc.ClientConn
	chunkSize  uint
	logger     *zerolog.Logger
}

func NewClient(cfg *ClientConfig) (*Client, error) {
	// Set up a connection to the server.
	conn, err := grpc.Dial(cfg.Address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	logger := zerolog.New(os.Stderr).With().Str("role", "client").Timestamp().
		Caller().Logger()

	return &Client{
		address:   cfg.Address,
		chunkSize: 1 << 13,
		client:    shsched.NewShschedClient(conn),
		logger:    &logger,
	}, nil
}

func (c Client) Close() error {
	c.logger.Info().Msgf("Close client connection")
	return c.connection.Close()
}

func (c *Client) GetInfo(ctx context.Context) (*shsched.NodeInfo, error) {
	c.logger.Debug().Msgf("call GetInfo")
	return c.client.GetInfo(ctx, &shsched.Empty{})
}

// type Recipe struct {
// 	ExecFilePath string `json:"execFilePath"`
// 	Run          string `json:"run"`
// }

func (c *Client) Exec(
	ctx context.Context,
	recipeFilePathRaw string,
) (err error) {
	// recipe, err := os.Open(recipeFilePath)
	// if err != nil {
	// 	c.logger.Error().Err(err).Msg("Open recipeFilePath error")
	// 	return err
	// }
	// defer recipe.Close()
	recipeFilePath := filepath.Clean(recipeFilePathRaw)
	recipeFileDir, _ := filepath.Split(recipeFilePath)
	// if err != nil {
	// 	return err
	// }
	// panic(recipeFileDir)

	recipeFile, err := os.Open(recipeFilePath)
	if err != nil {
		c.logger.Error().Err(err).Msg("Open recipeFile error")
		return err
	}

	recipeBytes, err := ioutil.ReadAll(recipeFile)
	if err != nil {
		c.logger.Error().Err(err).Msg("Read recipeFilePath error")
		return err
	}

	recipe := &types.Recipe{}
	err = json.Unmarshal(recipeBytes, recipe)
	if err != nil {
		c.logger.Error().Err(err).Msg("Unmarshal recipeFile from JSON error")
		return err
	}

	file, err := os.Open(filepath.Join(recipeFileDir, recipe.ExecFile))
	if err != nil {
		c.logger.Error().Err(err).Msg("Open recipeFilePath error")
		return err
	}
	defer file.Close()

	// Open a stream-based connection with the
	// gRPC server
	stream, err := c.client.Exec(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("Exec error")
		return err
	}

	// Start timing the execution
	// stats.StartedAt = time.Now()

	// Prepare first chunk

	// Send first chunk
	err = stream.Send(&shsched.Chunk{
		Content: recipeBytes,
	})
	if err != nil {
		c.logger.Error().Err(err).Msg("send first chunk error")
		return err
	}

	buf := make([]byte, c.chunkSize)
	writing := true

	for writing {
		n, err := file.Read(buf)
		if n > 0 {

		}
		if err != nil {
			if err == io.EOF {
				writing = false
				continue
			}

			c.logger.Error().Err(err).Msg("Exec error")
			return err
		}

		err = stream.Send(&shsched.Chunk{
			Content: buf[:n],
		})
		if err != nil {
			c.logger.Error().Err(err).Msg("send chunk error")
			return err
		}
	}

	// keep track of the end time so that we can take the elapsed
	// time later
	// stats.FinishedAt = time.Now()

	// close
	status, err := stream.CloseAndRecv()
	if err != nil {
		c.logger.Error().Err(err).Msg("Exec error")
		return err
	}
	fmt.Println("!!!", status.Message)

	return nil
}
