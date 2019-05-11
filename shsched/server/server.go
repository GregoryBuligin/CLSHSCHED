package shsched

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"google.golang.org/grpc"

	"github.com/rs/zerolog"

	types "shsched/shsched"
)

type ServerConfig struct {
	Port string
}

type Server struct {
	port     string
	server   *grpc.Server
	logger   *zerolog.Logger
	TaskChan chan types.Task
}

func NewServer(cfg *ServerConfig) *Server {
	logger := zerolog.New(os.Stderr).With().Str("role", "server").Timestamp().
		Caller().Logger()

	server := &Server{
		port:     cfg.Port,
		server:   grpc.NewServer(),
		logger:   &logger,
		TaskChan: make(chan types.Task, 10),
	}

	RegisterShschedServer(server.server, server)

	return server
}

func (s *Server) Serve() (err error) {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return err
	}

	s.logger.Info().Msgf("Server starts on port: %s", s.port)
	if err = s.server.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (s Server) GetInfo(context.Context, *Empty) (*NodeInfo, error) {
	s.logger.Debug().Msg("call GetInfo")
	return &NodeInfo{CPU: 90}, nil
}

func (s Server) Exec(stream Shsched_ExecServer) (err error) {
	s.logger.Debug().Msg("call Exec")

	dir, err := ioutil.TempDir("", "exec")
	if err != nil {
		s.logger.Error().Err(err).Msg("create TempDir error")
		return err
	}
	// defer os.RemoveAll(dir)

	firstChunk, err := stream.Recv()
	if err != nil {
		s.logger.Error().Err(err).Msg("recv first chunk error")
		return err
	}

	recipe := &types.Recipe{}
	err = json.Unmarshal(firstChunk.Content, recipe)
	if err != nil {
		s.logger.Error().Err(err).Msg("Unmarshal firstChunk from JSON error")
		return err
	}

	folderPath := fmt.Sprintf("%s/", dir)
	os.MkdirAll(folderPath, os.ModePerm)

	filePath := filepath.Join(folderPath, recipe.ExecFile)
	file, err := os.Create(filePath)
	if err != nil {
		s.logger.Error().Err(err).Msg("Create exec file error")
		return err
	}
	defer file.Close()

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				goto END
			}

			s.logger.Error().Err(err).Msg("GetInfo error")
			return err
		}

		if _, err = file.Write(chunk.Content); err != nil {
			s.logger.Error().Err(err).Msg("GetInfo error")
			return err
		}
	}

END:

	scriptPath := filepath.Join(folderPath, recipe.ExecFile)

	script := strings.Replace(recipe.Script, recipe.ExecFile, scriptPath, 1)
	cmd := exec.Command("/bin/sh", "-c", script)
	cmd.Dir = folderPath

	s.TaskChan <- types.Task{
		CMD: *cmd,
		Dir: dir,
	}

	// once the transmission finished, send the
	// confirmation if nothign went wrong
	err = stream.SendAndClose(&ExecResponse{
		Message: "Success task upload",
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("GetInfo error")
		return err
	}

	return nil
}

func (s Server) Ret(context.Context, *ExecOutput) (*Empty, error) {
	return nil, nil
}
