package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"log"
	api2 "os-project/api"
	"os-project/config"
)

func main() {
	schedulerConfig := config.GetSchedulerConfig()
	schedulerHandler := api2.NewSchedulerHandlerImpl(schedulerConfig)

	app := fiber.New()
	app.Use(logger.New())
	app.Use(recover.New())

	api := app.Group("/api")

	v1 := api.Group("/v1")
	{
		v1.Get("/fcfs", schedulerHandler.FirstComeFirstServe)
		v1.Get("/rr", schedulerHandler.RoundRobin)
		v1.Get("sjf", schedulerHandler.ShortestJobFirst)
		v1.Get("/mlfq", schedulerHandler.MultilevelFeedbackQueue)
		v1.Get("/all", schedulerHandler.AllAlgorithms)
	}

	log.Fatalln(app.Listen(fmt.Sprintf(":%d", schedulerConfig.Port)))
}
