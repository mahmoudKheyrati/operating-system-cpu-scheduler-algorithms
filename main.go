package main

import (
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	app := fiber.New()
	api := app.Group("/api")

	v1 := api.Group("/v1")
	{
		v1.Get("/fcfs")
		v1.Get("/rr")
		v1.Get("sjf")
		v1.Get("/mlfq")
		v1.Get("/all")
	}



	log.Fatalln(app.Listen(":9095"))
}
