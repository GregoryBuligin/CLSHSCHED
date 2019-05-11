package main

import (
	"context"
	"log"

	shschedClient "shsched/shsched/client"
)

const (
	address = "localhost:8000"
)

func main() {
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()

	ctx := context.Background()

	client, err := shschedClient.NewClient(&shschedClient.ClientConfig{
		Address: address,
	})
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}

	err = client.Exec(ctx, "prepared/Recipe.json")
	if err != nil {
		log.Fatalf("GetInfo: %v", err)
	}
	// log.Printf("resp: %s", resp)

	// resp, err := client.GetInfo(ctx)
	// if err != nil {
	// 	log.Fatalf("GetInfo: %v", err)
	// }
	// log.Printf("resp: %s", resp)
}
