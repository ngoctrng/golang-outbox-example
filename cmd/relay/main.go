package main

import (
	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"os"
	"os/signal"
	"outbox/database"
	"outbox/queue"
	"outbox/shared"
	"syscall"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("loading env file: ", err)
	}

	db, err := database.NewConnection()
	if err != nil {
		log.Fatal("error connecting to db")
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

	jobProcessor := shared.OutboxProcesser{
		DB:      db,
		Channel: ch,
		Queue:   q,
	}

	c := cron.New()
	_, err = c.AddFunc("@every 10s", jobProcessor.HandleOutboxMessage)
	if err != nil {
		log.Fatal("register handler error", err)
	}
	log.Println("Start processing outbox messages")
	c.Start()
	defer c.Stop()

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