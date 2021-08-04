package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"log"
	"outbox/customer"
	"outbox/database"
	"outbox/shared"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("loading env file: ", err)
	}

	db, err := database.NewConnection()
	if err != nil {
		log.Fatal("error connecting to db")
	}

	if err := db.AutoMigrate(&customer.Customer{}, &shared.OutBoxMessage{}); err != nil {
		log.Fatal("migrate error - ", err)
	}

	customerHandler := customer.Handler{DB: db}

	app := fiber.New()

	app.Use(logger.New())

	app.Post("/customers", customerHandler.Add)

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
