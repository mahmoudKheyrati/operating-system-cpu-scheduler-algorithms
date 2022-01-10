package api

import "github.com/gofiber/fiber/v2"

type SchedulerHandler interface {
	FirstComeFirstServe(ctx *fiber.Ctx)
	RoundRobin(ctx *fiber.Ctx)
	ShortestJobFirst(ctx *fiber.Ctx)
	MultilevelFeedbackQueue(ctx *fiber.Ctx)
	AllAlgorithms(ctx *fiber.Ctx)
}
