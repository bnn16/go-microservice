
# Go-Microservices

Building Scalable Microservices in Golang: Leveraging gRPC, RabbitMQ, and SQLite for Efficient Messaging and State Management with RESTful Automation


## Architecture

The initial idea after a bit of reasearch was the following: 

[![thumbnail-microgo.png](https://i.postimg.cc/CMmrh0sZ/thumbnail-microgo.png)](https://postimg.cc/Lhg3Dr7R)

A user sends a request to a rest api that invokes a gRPC. However that was a mistake, since the idea of the project was to be async, instead of directly calling the method from the restapi.

[![microgo-drawio2.png](https://i.postimg.cc/85sJLLxg/microgo-drawio2.png)](https://postimg.cc/dkPVYZq5)

Then after some more reasearch I figured out that the best way implement the RabbitMQ and gRPC was the following: 
API --> RabbitMQ Queue that then on goes to call the Microservice, and then they communicate between each other via gRPC.


[![microgo-drawio25.png](https://i.postimg.cc/sgDW4BBm/microgo-drawio25.png)](https://postimg.cc/zV63JGFH)

After getting my hands dirty with the RabbitMQ, I figured out I would need a way to call the callback function from the previous diagram, so I stumbled on the Request-reply pattern seen here : 
[![Screenshot-2023-12-14-at-14-25-02.png](https://i.postimg.cc/c41CVps4/Screenshot-2023-12-14-at-14-25-02.png)](https://postimg.cc/PNcdDFpB)


## ProtoFile

This is the initial proto file, that I used to build the 2 microservices, allow them to communicate. They have a function called AddItem, that when it's called it needs a message(ItemRequest) and it returns the message(ItemResponse)

```bash
syntax = "proto3";
option go_package = "grpc/ms/pb";

service MyService {
    rpc AddItem(ItemRequest) returns (ItemResponse) {}
}

message ItemRequest {
    string name = 1;
}

message ItemResponse {
    int64 id = 1;
    string name = 2;
}
```
    
## Roadmap

- Additional browser support

- Add more integrations


## Microservice2 2 (DB Helper)

I first had to implement the AddItem function that we previously defined in the protofile

*I'm just showing code snippets, if you want to find the full file it's server2/main.go*

```go
// function to add items to the database (CREATE)
func (s *server) AddItem(ctx context.Context, req *pb.ItemRequest) (*pb.ItemResponse, error) {

	//insert the item into the database
	result, err := s.db.Exec("INSERT INTO items(name) VALUES(?)", req.Name)
	//check for errors, due to sqlite instance not running?
	if err != nil {
		return nil, fmt.Errorf("failed to add item: %v", err)
	}

	//get the id of the last inserted item
	id, _ := result.LastInsertId()

	return &pb.ItemResponse{
		Id:   id,
		Name: req.Name,
	}, nil
}
```

Afterwards I craete the db as well as the gRPC server
```go
listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

    s := grpc.NewServer()
	pb.RegisterMyServiceServer(s, &server{db: db})

```


## Microservice2 1 (Customer and Producer)

With this Microservice, I had to implement a Customer as well as a Producer of requests


*I'm just showing code snippets, if you want to find the full file it's server2/main.go*

The following part is for the gRPC server - 
```go
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
```

And for the RabbitMQ Queue
```go
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
```
## restAPI


First I establish a connection to the RabbitMQ queue
```go
rabbitConn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	rabbitChannel, err := rabbitConn.Channel()
	if err != nil {
		log.Fatalf("failed to open a RabbitMQ channel: %v", err)
	}
	defer rabbitChannel.Close()

	queue, err := rabbitChannel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to declare a RabbitMQ queue: %v", err)
	}
	log.Printf("RabbitMQ queue declared: %v", queue)
	defer rabbitChannel.Close()
```

Then I made an endpoint
```go
e.POST("/add-item", func() error{...})
```
That was responsible for creating a CorrelationId as well as publishing the request/value of the request to the Queue, that will be picked up by the worker/server.

It returns a CorrelationId, since we don't know when the queue will be done. So a user can access the updated information, once it's processed
```go
response := map[string]interface{}{
			"message": "Item added and sent to RabbitMQ",
			"corrId":  corrId,
		}
return c.JSON(http.StatusOK, response)
```


## Tech Stack

**RestAPI:** Echo
**Queues:** RabbitMQ
**Server:** gRPC




## Resources 

 - [RabbitMQ Docs/Tutorials](https://www.rabbitmq.com/tutorials/tutorial-six-python.html)
 - [gRPC docs](https://grpc.io/docs/languages/go/basics/)
 - [RabbitMQ : Message Queues for beginners](https://www.youtube.com/watch?v=hfUIWe1tK8E&t=210s)
  - [Where should you use gRPC? And where NOT to use it!](https://www.youtube.com/watch?v=4SuFtQV8RCk&t=11s)
  - [Microservices communication using gRPC Protocol](https://medium.com/javarevisited/microservices-communication-using-grpc-protocol-dc3a2f8b648d)
  - [Scalable Microservice Architecture Using RabbitMQ RPC](https://medium.com/swlh/scalable-microservice-architecture-using-rabbitmq-rpc-d07fa8faac32)  

## Things that I couldn't do

**Issue with Correlation ID for Client Responses**

The setup for the Correlation ID, aimed at retrieving client responses post-queue, encountered a problem. Initially, I managed to get it to work smoothly, allowing clients to access their responses once the queue finished processing.

However, subsequent attempts faced a glitch. Despite troubleshooting efforts, it consistently broke after the initial success. Due to time constraints, I couldn't pinpoint the root cause and resolve the issue completely within the project's timeframe

[![image.png](https://i.postimg.cc/zGG4Tjqf/image.png)](https://postimg.cc/SJHZhc20)




## Deployment

I wasn't able to make the dockerfiles in time, since I had to reasearch how everything works, including golang...


