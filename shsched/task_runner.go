package shsched

import (
	"fmt"
	"os"
)

func StartRunner(server *Server, semaphore chan uint) {
	for task := range server.TaskChan {
		semaphore <- 1
		go func(task Task, semaphore chan uint) {
			defer func() {
				<-semaphore
				os.RemoveAll(task.Dir)
			}()

			out, err := task.CMD.Output()
			if err != nil {
				panic(err)
			}
			fmt.Println("RUN HERE!!!")

			fmt.Printf("OUTPUT for %s:\t%s\n", task.RetAddress, string(out))
			server.CompleteTaskOutputChan <- Output{
				RetAddress: task.RetAddress,
				Output:     string(out),
			}
		}(task, semaphore)
	}
}
