package main

import (
	"log"

	"shsched/shsched"
)

// const port = "8000"

const port = "8001"

var semaphore = make(chan uint, 100)

func main() {
	// ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	// defer cancel()
	// list, _ := netscanner.Scan(ctx)
	// fmt.Printf("%+v\n", list)
	// panic(list)

	cfg := &shsched.ServerConfig{
		Port: port,
	}

	server, err := shsched.NewServer(cfg)
	if err != nil {
		log.Fatalf("%+v\n", err)
	}

	// panic(server.Port)

	go server.OutputWaiter()
	go server.SelectTask()
	go shsched.StartRunner(server, semaphore)

	// go func() {
	// 	time.Sleep(time.Second * 10)
	// 	myHost, myFirstPost, err := netscanner.ScanMyIP(
	// 		context.Background(),
	// 		"127.0.0.1",
	// 	)
	// 	if err != nil {
	// 		log.Fatalf("ScanMyIP: %v", err)
	// 	}
	//
	// 	client, err := shsched.NewClient(&shsched.ClientConfig{
	// 		Address:    fmt.Sprintf("%s:%d", myHost, myFirstPost),
	// 		ServerPort: port,
	// 	})
	// 	if err != nil {
	// 		log.Fatalf("NewClient: %v", err)
	// 	}
	//
	// 	_, err = client.SchedTask(context.Background(), "prepared/Recipe.json")
	// 	if err != nil {
	// 		log.Fatalf("SchedTask: %v", err)
	// 	}
	// }()

	if err := server.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
