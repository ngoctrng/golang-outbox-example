package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"gorm.io/datatypes"
	"io"
	"log"
	"os"
	"os/signal"
	"outbox/queue"
	"syscall"
)

type OutboxEvent struct {
	ID        string         `json:"id"`
	EventName string         `json:"event_name"`
	Payload   datatypes.JSON `json:"payload"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("loading env file: ", err)
	}

	conn, err := queue.CreateConnection()
	if err != nil {
		log.Fatal(err)
	}
	defer closeConnection(conn)

	ch, err := queue.CreateChannel(conn)
	if err != nil {
		log.Fatal(err)
	}
	defer closeConnection(ch)

	q, err := ch.QueueDeclare(
		"outbox",
		false, false, false, false, nil)

	messages, err := ch.Consume(
		q.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)

	// Run in background
	go func() {
		log.Printf("Comsuming queue [%s] \n", q.Name)
		for m := range messages {
			var evt OutboxEvent
			if err := json.Unmarshal(m.Body, &evt); err != nil {
				log.Println("Handle message error: ", string(m.Body))
				log.Println("ERR:", err)
				continue
			}
			log.Printf("Handling [%s] - Payload: '%s'", evt.EventName, evt.Payload)
		}
	}()

	// Wait for terminate signal
	kill := make(chan os.Signal, 1)
	signal.Notify(kill, syscall.SIGINT, syscall.SIGTERM)
	<-kill
}

func closeConnection(c io.Closer) {
	err := c.Close()
	if err != nil {
		log.Println(err)
	}
}
