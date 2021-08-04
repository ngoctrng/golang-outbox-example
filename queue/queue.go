package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"os"
)

func CreateConnection() (*amqp.Connection, error) {
	ampqURL := fmt.Sprintf(
		"amqp://%s:%s@%s:%s",
		os.Getenv("RABBITMQ_USER"), os.Getenv("RABBITMQ_PASS"),
		os.Getenv("RABBITMQ_HOST"), os.Getenv("RABBITMQ_PORT"))
	return amqp.Dial(ampqURL)
}

func CreateChannel(conn *amqp.Connection) (*amqp.Channel, error) {
	return conn.Channel()
}
