package main

import (
	"flag"
	"log"

	"github.com/arturoeanton/99F/pkg/jsonschema"
	"github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

//var store *session.Store

func middlewareAuth0(c *fiber.Ctx) error {
	return c.Next()
}

func main() {

	addr := flag.String("addr", ":9090", "http service address")
	flag.Parse()

	//store = session.New()

	fiberApp := fiber.New()
	fiberApp.Use(logger.New())
	fiberApp.Use(requestid.New())
	fiberApp.Use(recover.New())
	fiberApp.Static("/site/", "./site")
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/site/index.html")
	})

	fiberApp.Get("/schema/:name", func(c *fiber.Ctx) error {
		name := c.Params("name", "demo")
		schema, _ := jsonschema.GetSchema(name)
		return c.JSON(schema)
	})

	fiberApp.Post("/resource/:name", func(c *fiber.Ctx) error {
		name := c.Params("name", "demo")
		elem, err := jsonschema.Bind(name, c.Body())
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(elem)
	})

	log.Fatal(fiberApp.Listen(*addr))
}
