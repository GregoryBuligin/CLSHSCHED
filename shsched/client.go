package shsched

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
)

const (
	address = "localhost:8000"
)

type ClientConfig struct {
	Address    string
	ServerPort string
}

type Client struct {
	address    string
	client     ShschedClient
	connection *grpc.ClientConn
	chunkSize  uint
	logger     *zerolog.Logger

	serverPort string
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
		client:    NewShschedClient(conn),
		logger:    &logger,
	}, nil
}

func (c Client) Close() error {
	c.logger.Info().Msgf("Close client connection")
	return c.connection.Close()
}

func (c *Client) GetInfo(ctx context.Context) (*NodeInfo, error) {
	c.logger.Debug().Msgf("call GetInfo")
	return c.client.GetInfo(ctx, &Empty{})
}

func (c *Client) Exec(
	ctx context.Context,
	recipeFilePathRaw string,
	retAddress string,
) (err error) {
	recipeFilePath := filepath.Clean(recipeFilePathRaw)
	recipeFileDir, _ := filepath.Split(recipeFilePath)

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

	recipe := &Recipe{}
	err = json.Unmarshal(recipeBytes, recipe)
	if err != nil {
		c.logger.Error().Err(err).Msg("Unmarshal recipeFile from JSON error")
		return err
	}

	recipe.RetAddress = retAddress

	// fmt.Println("recipe.RetAddress", recipe.RetAddress)
	// panic("***")

	file, err := os.Open(filepath.Join(recipeFileDir, recipe.ExecFile))
	if err != nil {
		c.logger.Error().Err(err).Msg("Open recipeFilePath error")
		return err
	}
	defer file.Close()

	// Open a stream-based connection with the gRPC server
	stream, err := c.client.Exec(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("Exec error")
		return err
	}

	newRecipeBytes, err := json.Marshal(recipe)
	if err != nil {
		c.logger.Error().Err(err).Msg("JSON Marshal error")
		return err
	}

	// Send first chunk
	err = stream.Send(&Chunk{
		Content: newRecipeBytes,
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

		err = stream.Send(&Chunk{
			Content: buf[:n],
		})
		if err != nil {
			c.logger.Error().Err(err).Msg("send chunk error")
			return err
		}
	}

	// Close stream
	status, err := stream.CloseAndRecv()
	if err != nil {
		c.logger.Error().Err(err).Msg("Exec error")
		return err
	}
	fmt.Println("!!!", status.Message)

	return nil
}

func (c *Client) SchedTask(
	ctx context.Context,
	recipeFilePathRaw string,
) (*Empty, error) {
	recipe := &RecipeMsg{
		RecipeFilePath: recipeFilePathRaw,
		Port:           c.serverPort,
	}

	fmt.Printf("%+v\n", recipe)

	_, err := c.client.SchedTask(ctx, recipe)
	if err != nil {
		c.logger.Error().Err(err).Msg("SchedTask error")
		return nil, err
	}

	return nil, nil
}

func (c *Client) Ret(ctx context.Context, in *ExecOutput) (*Empty, error) {
	return c.client.Ret(ctx, in)
}
