package shsched

import (
	"fmt"
	"os"
)

func StartRunner(server *Server, semaphore chan uint) {
	for task := range server.TaskChan {
		semaphore <- 1
		go func(task Task, semaphore chan uint) {
			fmt.Println("SERVER::", server.Port)
			defer func() {
				<-semaphore
				os.RemoveAll(task.Dir)
			}()

			out, err := task.CMD.Output()
			if err != nil {
				panic(err)
			}

			fmt.Println("string(out) ||||||||||||||||||||", string(out))
			server.CompleteTaskOutputChan <- Output{
				RetAddress: task.RetAddress,
				Output:     string(out),
			}
		}(task, semaphore)
	}
}
