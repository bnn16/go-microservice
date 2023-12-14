package main

import (
	"context"
	"encoding/json"
	"log"
	"net"

	pb "grpc/ms/pb"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
)

const (
	port       = ":50052"                // port for the intermediate microservice
	address    = "localhost:50051"       // address of the existing gRPC server
	rabbitURL  = "amqp://guest:guest@localhost:5672/"
	queueName  = "items_queue"
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

	// start RabbitMQ consumer
	go func() {
		connRabbit, err := amqp.Dial(rabbitURL)
		if err != nil {
			log.Fatalf("failed to connect to RabbitMQ: %v", err)
		}
		defer connRabbit.Close()

		ch, err := connRabbit.Channel()
		if err != nil {
			log.Fatalf("failed to open a channel: %v", err)
		}
		defer ch.Close()

		q, err := ch.QueueDeclare(
			queueName, // name
			false,     // durable
			false,     // delete when unused
			false,     // exclusive
			false,     // no-wait
			nil,       // arguments
		)
		if err != nil {
			log.Fatalf("failed to declare a queue: %v", err)
		}

		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			log.Fatalf("failed to register a consumer: %v", err)
		}

		for msg := range msgs {
			// using it this way because the message is a JSON object
			var receivedItem map[string]interface{}
			err := json.Unmarshal(msg.Body, &receivedItem)
			if err != nil {
				log.Printf("Error decoding message: %v", err)
				continue
			}

			itemReq := &pb.ItemRequest{
				Name:"",
			}

			if name, ok := receivedItem["Name"].(string); ok {
				itemReq.Name = name
			} else {
				log.Println("Error: Missing or invalid 'Name' field in received item")
				continue
			}

			// check if the Name field is present and is a string
			if(itemReq.Name == ""){
				log.Println("Error: Missing or invalid 'Name' field in received item")
				continue
			}
		
			// forward the AddItem request to the existing gRPC server
			_, err = client.AddItem(context.Background(), itemReq)
			if err != nil {
				log.Printf("Error calling AddItem: %v", err)
				continue
			}
		}
	}()

	if err := intermediateServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
