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

	jsonschema.InitDB()
	jsonschema.CreateBuckets("./schemas/")

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
		element, err := jsonschema.Bind(name, c.Body())
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		element, err = jsonschema.Create(element, name)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(element)
	})

	fiberApp.Put("/resource/:name/:id", func(c *fiber.Ctx) error {
		name := c.Params("name", "demo")
		id := c.Params("id", "")
		element, err := jsonschema.Bind(name, c.Body())
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		element, err = jsonschema.Replace(element, id, name)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(element)
	})

	fiberApp.Get("/resource/:name/:id", func(c *fiber.Ctx) error {
		name := c.Params("name", "demo")
		id := c.Params("id", "")

		element, err := jsonschema.GetElementByID(id, name)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(element)
	})

	fiberApp.Get("/resource/:name", func(c *fiber.Ctx) error {
		name := c.Params("name")
		filter := c.Query("filter")
		startIndex := c.Query("startIndex")
		count := c.Query("count")
		sortBy := c.Query("sortBy")
		sortOrder := c.Query("sortOrder")
		result, err := jsonschema.Search(name, filter, startIndex, count, sortBy, sortOrder)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(result)
	})

	fiberApp.Delete("/resource/:name/:id", func(c *fiber.Ctx) error {
		name := c.Params("name", "demo")
		id := c.Params("id", "")
		err := jsonschema.Remove(id, name)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(fasthttp.StatusNoContent).SendString("")
	})

	log.Fatal(fiberApp.Listen(*addr))
}
