package main

import (
	"flag"
	"log"

	"github.com/arturoeanton/99F/pkg/jsonschema"
	"github.com/arturoeanton/99F/pkg/runnerjs"
	"github.com/arturoeanton/gocommons/utils"
	"github.com/valyala/fasthttp"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

// podman run --rm -it --hostname localhost -p 15672:15672 -p 5672:5672 rabbitmq:3-management
// podman run -d --name db -p 8091-8094:8091-8094 -p 11210:11210 couchbase

func main() {

	addr := flag.String("addr", ":3000", "http service address")
	flag.Parse()

	jsonschema.InitDB()
	jsonschema.CreateBuckets("./entities/")

	fiberApp := fiber.New()
	fiberApp.Use(logger.New())
	fiberApp.Use(requestid.New())
	fiberApp.Use(recover.New())
	fiberApp.Static("/site/", "./site")
	fiberApp.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/site/index.html")
	})

	fiberApp.Get("/schema/:name", func(c *fiber.Ctx) error {
		name := c.Params("name", "")
		schema, err := jsonschema.GetSchema(name)
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(schema)
	})

	fiberApp.Post("/validate/:name", func(c *fiber.Ctx) error {
		name := c.Params("name", "")
		_, err := jsonschema.Bind(name, c.Body())
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		return c.SendStatus(fasthttp.StatusNoContent)
	})

	fiberApp.Post("/resource/:name", func(c *fiber.Ctx) error {
		name := c.Params("name", "")
		schema, err := jsonschema.GetSchema(name)
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		element, err := jsonschema.Bind(name, c.Body())
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		element, id, err := jsonschema.Create(element, name)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if utils.Exists("entities/" + name + "/constructor.js") {
			go runnerjs.Run(c, element, schema, id, name, "constructor", "_constructor")
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
		schema, err := jsonschema.GetSchema(name)
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		element, err := jsonschema.Remove(id, name)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if utils.Exists("entities/" + name + "/destructor.js") {
			go runnerjs.Run(c, element, schema, id, name, "destructor", "_destructor")
		}
		return c.Status(fasthttp.StatusNoContent).SendString("")
	})

	fiberApp.Get("/form/:name", func(c *fiber.Ctx) error {
		name := c.Params("name", "")
		html, err := jsonschema.Form(name)
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
		return c.SendString(html)
	})

	fiberApp.Get("/resource/:name/:id/:method", func(c *fiber.Ctx) error {
		name := c.Params("name", "")
		id := c.Params("id", "")
		method := c.Params("method", "")

		schema, err := jsonschema.GetSchema(name)
		if err != nil {
			return c.Status(fasthttp.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		element, err := jsonschema.GetElementByID(id, name)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}

		err = runnerjs.Run(c, element, schema, id, name, method, method)
		if err != nil {
			return c.Status(fasthttp.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return nil
	})

	log.Fatal(fiberApp.Listen(*addr))
}
