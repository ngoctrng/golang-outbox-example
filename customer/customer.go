package customer

import (
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"outbox/shared"
	"time"
)

type Customer struct {
	ID        string `json:"id" gorm:"id,primarykey"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Handler struct {
	DB *gorm.DB
}

func (h *Handler) Add(c *fiber.Ctx) error {

	var customer Customer
	if err := c.BodyParser(&customer); err != nil {
		return err
	}
	customer.ID = uuid.NewString()
	customer.CreatedAt = time.Now()

	err := h.DB.Transaction(func(tx *gorm.DB) error {
		b, err := json.Marshal(customer)
		if err != nil {
			return err
		}

		customerCreatedEvent := shared.OutBoxMessage{
			ID:          uuid.NewString(),
			EventName:   "CustomerCreated",
			Payload:     datatypes.JSON(b),
			IsProcessed: false,
		}

		if err := tx.FirstOrCreate(&customer).Error; err != nil {
			return err
		}

		if err := tx.Create(&customerCreatedEvent).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
