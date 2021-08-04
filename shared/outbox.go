package shared

import (
	"encoding/json"
	"github.com/streadway/amqp"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"log"
	"time"
)

type OutBoxMessage struct {
	ID          string         `gorm:"id" json:"id"`
	EventName   string         `gorm:"event_name" json:"event_name"`
	Payload     datatypes.JSON `gorm:"payload" json:"payload"`
	IsProcessed bool           `gorm:"is_processed" json:"is_processed"`
}

type OutboxProcesser struct {
	DB      *gorm.DB
	Channel *amqp.Channel
	Queue   amqp.Queue
}

func (p *OutboxProcesser) HandleOutboxMessage() {
	messages := make([]OutBoxMessage, 0)
	err := p.DB.
		Where("is_processed = ?", false).
		Find(&messages).Error
	if err != nil {
		log.Println("query outbox messages error: ", err)
		return
	}

	// no waiting message
	if len(messages) == 0 {
		return
	}

	// Publish each message.
	// If success, add to processed slice
	processedID := make([]string, 0)
	for _, m := range messages {
		b, err := json.Marshal(m)
		if err != nil {
			continue
		}

		// publish message to a queue
		if err := p.publishMessage(b); err != nil {
			log.Println("publish outbox message error: ", err)
			continue
		}

		processedID = append(processedID, m.ID)
	}

	// Update processed messages in database
	err = p.DB.Model(&OutBoxMessage{}).
		Where("id IN ?", processedID).
		UpdateColumn("is_processed", true).Error
	if err != nil {
		log.Println("update outbox error: ", err)
		return
	}

	log.Println("Published messages:", processedID)
}

func (p *OutboxProcesser) publishMessage(body []byte) error {
	return p.Channel.Publish(
		"",
		p.Queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Timestamp:   time.Now(),
			Body:        body,
		},
	)
}
