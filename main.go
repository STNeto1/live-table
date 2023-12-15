package main

import (
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/stneto1/htmx-webcomponents/pkg"
)

func main() {
	conn := pkg.CreateConnection(false) // boolean for whether to log queries
	container := pkg.NewContainer(conn)

	app := fiber.New(fiber.Config{})
	// app.Use(logger.New())
	app.Static("/public", "./public")

	app.Get("/", container.IndexHandler)
	app.Post("/reseed", container.ReseedHandler)
	app.Get("/ws", websocket.New(container.RecordsWsHandler, websocket.Config{}))

	go container.RunHub()

	app.Listen(":3000")
}
