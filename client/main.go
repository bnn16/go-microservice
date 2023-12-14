package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"

	pb "grpc/ms/pb"
)

const (
	address = "localhost:50052"
)

// function to add items to the database (CREATE)
func addItem(client pb.MyServiceClient) {
	var itemName string

	fmt.Print("Enter the item name to add: ")
	fmt.Scanln(&itemName)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	response, err := client.AddItem(ctx, &pb.ItemRequest{Name: itemName})
	if err != nil {
		log.Fatalf("could not add item via intermediate service: %v", err)
	}

	log.Printf("Item added via intermediate service. ID: %d, Name: %s", response.Id, response.Name)
}

func main() {
	// set up a connection to the intermediate gRPC server
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect to intermediate service: %v", err)
	}
	defer conn.Close()

	// create a gRPC client for the intermediate MyService
	client := pb.NewMyServiceClient(conn)

	// add an item via the intermediate gRPC server
	addItem(client)
}
