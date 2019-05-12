package shsched

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"

	"github.com/rs/zerolog"

	"shsched/netscanner"
)

type ServerConfig struct {
	Port string
}

type Server struct {
	networkIP    string
	networkHosts netscanner.Hosts

	Port                   string
	server                 *grpc.Server
	logger                 *zerolog.Logger
	TaskChan               chan Task
	WaitTaskChan           chan Input
	CompleteTaskOutputChan chan Output

	// Client       *Client
}

func NewServer(cfg *ServerConfig) (*Server, error) {
	logger := zerolog.New(os.Stderr).With().Str("role", "server").Timestamp().
		Caller().Logger()

	myIP, err := netscanner.ExternalIP()
	if err != nil {
		return nil, err
	}

	myPort, err := strconv.Atoi(cfg.Port)
	if err != nil {
		return nil, err
	}

	networkHosts, err := netscanner.Scan(
		context.Background(),
		myIP,
		uint16(myPort),
	)
	if err != nil {
		return nil, err
	}

	if len(networkHosts) == 0 {
		logger.Warn().Msg("running nodes not found!")
	}

	fmt.Printf("!!!%+v!!!\n", len(networkHosts))

	server := &Server{
		networkIP:    myIP,
		networkHosts: networkHosts,

		Port:                   cfg.Port,
		server:                 grpc.NewServer(),
		logger:                 &logger,
		TaskChan:               make(chan Task, 100),
		WaitTaskChan:           make(chan Input, 100),
		CompleteTaskOutputChan: make(chan Output, 100),
		// Client:       client,
	}

	RegisterShschedServer(server.server, server)

	return server, nil
}

func (s *Server) getNewClient() (client *Client, err error) {
	var address string
	var port string

	// panic(len(s.networkHosts))

	if len(s.networkHosts) > 0 {
		for k, v := range s.networkHosts {
			// fmt.Println(k, s.networkIP)
			// panic(k == s.networkIP)
			if k == s.networkIP {
				continue
			}

			if len(v) > 0 {
				address = k
				port = strconv.Itoa(int(v[0]))
			} else {
				continue
			}

			// !!!!!!!!!!!!!
			// delete(s.networkHosts, k)
			break
		}
	} else {
		address = "127.0.0.1"
		port = s.Port
	}

	client, err = NewClient(&ClientConfig{
		Address:    fmt.Sprintf("%s:%s", address, port),
		ServerPort: s.Port,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (s *Server) Serve() (err error) {
	lis, err := net.Listen("tcp", ":"+s.Port)
	if err != nil {
		return err
	}

	s.logger.Info().Msgf("Server starts on port: %s", s.Port)
	if err = s.server.Serve(lis); err != nil {
		return err
	}

	return nil
}

func (s *Server) GetInfo(context.Context, *Empty) (*NodeInfo, error) {
	s.logger.Debug().Msg("call GetInfo")
	return &NodeInfo{CPU: 90}, nil
}

func (s *Server) Exec(stream Shsched_ExecServer) (err error) {
	s.logger.Debug().Msg("call Exec")

	host, port, err := getRetIPByContext(stream.Context())
	if err != nil {
		s.logger.Error().Err(err).Msg("getRetIPByContext error")
		return err
	}

	dir, err := ioutil.TempDir("", "exec")
	if err != nil {
		s.logger.Error().Err(err).Msg("create TempDir error")
		return err
	}

	firstChunk, err := stream.Recv()
	if err != nil {
		s.logger.Error().Err(err).Msg("recv first chunk error")
		return err
	}

	recipe := &Recipe{}
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

	s.TaskChan <- Task{
		CMD:        *cmd,
		Dir:        dir,
		RetAddress: recipe.RetAddress,
	}

	// Message: fmt.
	// 	Sprintf("Success task upload to %s:%s", s.networkIP, s.port),

	// once the transmission finished, send the
	// confirmation if nothign went wrong
	err = stream.SendAndClose(&ExecResponse{
		Message: fmt.Sprintf("Success task upload from %s:%s", host, port),
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("GetInfo error")
		return err
	}

	return nil
}

func (s *Server) SchedTask(ctx context.Context, in *RecipeMsg) (*Empty, error) {
	s.logger.Debug().Msg("run SchedTask")

	s.WaitTaskChan <- Input{
		RecipeFilePath: in.RecipeFilePath,
		RetAddress:     fmt.Sprintf("%s:%s", s.networkIP, s.Port),
	}
	// panic(fmt.Sprintf("%s:%s", s.networkIP, s.Port))
	return &Empty{}, nil
}

func (s *Server) OutputWaiter() {
	s.logger.Debug().Msg("run OutputWaiter")
	for output := range s.CompleteTaskOutputChan {
		client, err := s.getNewClient()
		if err != nil {
			s.logger.Error().Err(err).Msg("getNewClient error")
			panic(err)
		}
		defer client.Close()
		client.address = output.RetAddress
		// panic(client.address)

		_, err = client.Ret(context.Background(), &ExecOutput{
			Output: output.Output,
		})
		if err != nil {
			s.logger.Error().Err(err).Msg("getNewClient error")
			panic(err)
		}
	}
}

func (s *Server) Ret(ctx context.Context, in *ExecOutput) (*Empty, error) {
	s.logger.Debug().Msg("run Ret")
	fmt.Println("server output:", in.Output)

	return &Empty{}, nil
}

func (s *Server) SelectTask() (err error) {
	s.logger.Debug().Msg("run SelectTask")

	fmt.Println(s.networkIP, s.Port)
	// panic(">>>")

	for task := range s.WaitTaskChan {
		client, err := s.getNewClient()
		if err != nil {
			s.logger.Error().Err(err).Msg("getNewClient error")
			panic(err)
		}
		defer client.Close()

		// fmt.Println("TTT", client.serverPort)
		// retAddress := fmt.Sprintf("%s:%s", s.networkIP, task.Port)
		// panic("QQQ" + task.RetAddress)

		err = client.Exec(
			context.Background(),
			task.RecipeFilePath,
			task.RetAddress,
		)
		if err != nil {
			s.logger.Error().Err(err).Msgf("run SelectTask on %s", task)
		}
	}

	return nil
}

func getRetIPByContext(ctx context.Context) (string, string, error) {
	var err error
	pr, ok := peer.FromContext(ctx)
	if !ok {
		err = fmt.Errorf("getClinetIP, invoke FromContext() failed")
	}
	if pr.Addr == net.Addr(nil) {
		err = fmt.Errorf("getClientIP, peer.Addr is nil")
	}
	if err != nil {
		return "", "", err
	}

	hostPort := strings.Split(pr.Addr.String(), ":")
	if len(hostPort) != 2 {
		return "", "", errors.New("incorrect address")
	}

	host := hostPort[0]
	port := hostPort[1]

	return host, port, nil
}
