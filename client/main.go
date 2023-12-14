package main

import (
	"context"
	"log"

	"google.golang.org/grpc"

	pb "grpc/ms/pb"
)

const (
	address = "localhost:50051"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// Create a client instance of the service
	client := pb.NewMyServiceClient(conn)

	// Call the AddItem RPC
	item := &pb.ItemRequest{Name: "Sample Item"}
	response, err := client.AddItem(context.Background(), item)
	if err != nil {
		log.Fatalf("could not add item: %v", err)
	}

	// Print the response
	log.Printf("Added Item: ID %d, Name %s", response.Id, response.Name)
}
