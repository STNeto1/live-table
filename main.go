package main

import (
	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v2"
	"github.com/stneto1/htmx-webcomponents/views"
)

func main() {
	app := fiber.New(fiber.Config{})
	// app.Use(logger.New())
	app.Static("/public", "./public")

	app.Get("/", func(c *fiber.Ctx) error {
		btn := views.RootLayout("Page Title")

		return render(btn, c)
	})

	app.Listen(":3000")
}

func render(component templ.Component, c *fiber.Ctx) error {
	c.Response().Header.SetContentType("text/html")
	return component.Render(c.Context(), c.Response().BodyWriter())
}
