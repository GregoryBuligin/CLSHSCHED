package main

// import (
// 	"shsched/shsched"
// )
//
// func getClient() {
// 	client := shsched.NewShschedClient(conn)
// }

// func main() {
// 	// Set up a connection to the server.
// 	conn, err := grpc.Dial(address, grpc.WithInsecure())
// 	if err != nil {
// 		log.Fatalf("did not connect: %v", err)
// 	}
// 	defer conn.Close()
// 	client := shsched.NewShschedClient(conn)
//
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
// 	resp, err := client.GetInfo(ctx, &shsched.Empty{})
// 	if err != nil {
// 		log.Fatalf("could not greet: %v", err)
// 	}
// 	log.Printf("Greeting: %s", resp)
// }
