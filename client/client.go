package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"shsched/netscanner"
	"shsched/shsched"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	address, err := netscanner.ExternalIP()
	if err != nil {
		log.Fatalf("ExternalIP: %v", err)
	}

	myHost, _, err := netscanner.ScanMyIP(ctx, address)
	if err != nil {
		log.Fatalf("ScanMyIP: %v", err)
	}

	client, err := shsched.NewClient(&shsched.ClientConfig{
		Address: fmt.Sprintf("%s:%d", myHost, 8001),
	})
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}

	_, err = client.SchedTask(context.Background(), "prepared/Recipe.json")
	if err != nil {
		log.Fatalf("SchedTask: %v", err)
	}

	// err = client.Exec(ctx, "prepared/Recipe.json")
	// if err != nil {
	// 	log.Fatalf("GetInfo: %v", err)
	// }
	// log.Printf("resp: %s", resp)

	// resp, err := client.GetInfo(ctx)
	// if err != nil {
	// 	log.Fatalf("GetInfo: %v", err)
	// }
	// log.Printf("resp: %s", resp)
}
