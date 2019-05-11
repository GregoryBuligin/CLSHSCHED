package main

import (
	"fmt"
	"log"
	"os"

	"shsched/netscanner"
	types "shsched/shsched"
	shschedServer "shsched/shsched/server"
)

const port = "8000"

var semaphore = make(chan uint, 100)

func main() {
	netscanner.Scan()

	cfg := &shschedServer.ServerConfig{
		Port: port,
	}

	server := shschedServer.NewServer(cfg)

	go func(server *shschedServer.Server) {
		for task := range server.TaskChan {
			semaphore <- 1
			go func(task types.Task, semaphore chan uint) {
				defer func() {
					<-semaphore
					os.RemoveAll(task.Dir)
				}()

				out, err := task.CMD.Output()
				if err != nil {
					panic(err)
				}
				// panic(string(out))
				fmt.Println(string(out))
			}(task, semaphore)
		}
	}(server)

	if err := server.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
