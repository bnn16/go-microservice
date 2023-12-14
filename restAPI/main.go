package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/streadway/amqp"
)

const (
    address   = "localhost:50052" // gRPC server address
    rabbitURL = "amqp://guest:guest@localhost:5672/"
    queueName = "items_queue"
)


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

	queue,err := rabbitChannel.QueueDeclare(queueName, false, false, false, false, nil)
	if err != nil {
		log.Fatalf("failed to declare a RabbitMQ queue: %v", err)
	}
	log.Printf("RabbitMQ queue declared: %v", queue)
	defer rabbitChannel.Close()


    e := echo.New()

	e.POST("/add-item", func(c echo.Context) error {
		var item map[string]interface{}
		if err := c.Bind(&item); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request payload")
		}

		messageBody, _ := json.Marshal(item)
		err = rabbitChannel.Publish("", queueName, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        messageBody,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to publish message to RabbitMQ")
		}

		return c.String(http.StatusOK, "Item added and sent to RabbitMQ")
	})


    log.Printf("REST API Server listening on port 8080")
    e.Logger.Fatal(e.Start(":8080"))
}
