package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/stneto1/htmx-webcomponents/pkg"
)

func main() {
	conn := pkg.CreateConnection(false) // boolean for whether to log queries
	container := pkg.NewContainer(conn)

	app := fiber.New(fiber.Config{})
	app.Use(logger.New())
	app.Static("/public", "./public")

	app.Get("/", container.IndexHandler)
	app.Post("/reseed", container.ReseedHandler)

	app.Listen(":3000")
}
