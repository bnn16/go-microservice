package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
)

const (
	address           = "localhost:50052" // gRPC server address
	rabbitURL         = "amqp://guest:guest@localhost:5672/"
	queueName         = "items_queue"
	responseQueueName = "items_response_queue"
)

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func main() {
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

	// responseChannel, err := rabbitConn.Channel()
	// if err != nil {
	// 	log.Fatalf("failed to open a RabbitMQ channel: %v", err)
	// }
	// defer responseChannel.Close()

	// responseQueue, err := responseChannel.QueueDeclare(responseQueueName, false, false, false, false, nil)
	// if err != nil {
	// 	log.Fatalf("failed to declare a RabbitMQ queue: %v", err)
	// }
	// log.Printf("RabbitMQ queue declared: %v", responseQueue)
	// defer rabbitChannel.Close()

	e := echo.New()

	e.POST("/add-item", func(c echo.Context) error {
		var item map[string]interface{}
		if err := c.Bind(&item); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
		}

		messageBody, _ := json.Marshal(item)
		corrId := randomString(32)

		err = rabbitChannel.Publish("", queueName, false, false, amqp.Publishing{
			ContentType:   "application/json",
			Body:          messageBody,
			CorrelationId: corrId,
			ReplyTo:       responseQueueName,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to publish message to RabbitMQ")
		}

		response := map[string]interface{}{
			"message": "Item added and sent to RabbitMQ",
			"corrId":  corrId,
		}

		return c.JSON(http.StatusOK, response)
	})

    //this is for the get-item endpoint, so the user can actually get the mutable item from the database
	// e.GET("/get-item/:id", func(c echo.Context) error {
	// 	id := c.Param("id")

	// 	msgs, err := responseChannel.Consume(responseQueueName, "", true, false, false, false, nil)
	// 	if err != nil {
	// 		return echo.NewHTTPError(http.StatusInternalServerError, "failed to register a consumer")
	// 	}

	// 	for msg := range msgs {
	// 		log.Printf("Received message: %s", msg.Body)
	// 		// Assuming message is JSON, unmarshal the message body
	// 		var item map[string]interface{}
	// 		if err := json.Unmarshal(msg.Body, &item); err != nil {
	// 			return echo.NewHTTPError(http.StatusInternalServerError, "failed to parse message from RabbitMQ")
	// 		}

	// 		// Check if the received item matches the requested ID
	// 		if receivedID, ok := item["id"].(string); ok && receivedID == id {
	// 			// Process the item
	// 			return c.JSON(http.StatusOK, item)
	// 		}
	// 		log.Printf("Received message does not match requested ID")
	// 	}

	// 	// If the loop finishes without finding the item, return an error or not found message
	// 	return echo.NewHTTPError(http.StatusNotFound, "item not found")
	// })

	log.Printf("REST API Server listening on port 8080")
	e.Logger.Fatal(e.Start(":8080"))
}
