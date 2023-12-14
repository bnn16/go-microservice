package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"

	pb "grpc/ms/pb"
)

const (
	port    = ":50052"               // port for the intermediate microservice
	address = "localhost:50051"      // address of the existing gRPC server
)

type intermediateService struct {
	pb.UnimplementedMyServiceServer
	client pb.MyServiceClient
}

func (s *intermediateService) AddItem(ctx context.Context, req *pb.ItemRequest) (*pb.ItemResponse, error) {
	// forward the AddItem request to the existing gRPC server
	response, err := s.client.AddItem(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func main() {
	// set up a connection to the existing gRPC server
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	// create a gRPC client for the existing MyService
	client := pb.NewMyServiceClient(conn)

	// set up the intermediate gRPC server
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// create the intermediate gRPC server
	intermediateServer := grpc.NewServer()
	pb.RegisterMyServiceServer(intermediateServer, &intermediateService{client: client})

	log.Printf("Intermediate Server listening on port %s", port)
	if err := intermediateServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
