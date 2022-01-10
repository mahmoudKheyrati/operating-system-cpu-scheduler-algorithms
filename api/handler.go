package api

import (
	"github.com/gofiber/fiber/v2"
	"os-project/config"
	"os-project/internal/requests"
	"os-project/internal/schedulers"
)

type SchedulerHandler interface {
	FirstComeFirstServe(ctx *fiber.Ctx) error
	RoundRobin(ctx *fiber.Ctx) error
	ShortestJobFirst(ctx *fiber.Ctx) error
	MultilevelFeedbackQueue(ctx *fiber.Ctx) error
	AllAlgorithms(ctx *fiber.Ctx) error
}
type SchedulerHandlerImpl struct {
	config *config.SchedulerConfig
}

func NewSchedulerHandlerImpl(config *config.SchedulerConfig) *SchedulerHandlerImpl {
	return &SchedulerHandlerImpl{config: config}
}

func (s *SchedulerHandlerImpl) FirstComeFirstServe(ctx *fiber.Ctx) error {
	var request *requests.ScheduleRequests
	if err := ctx.BodyParser(request); err != nil {
		ctx.JSON(fiber.Map{
			"error": "invalid request format",
		})
		return nil
	}
	response, err := schedulers.ScheduleFirstComeFirstServe(request)
	if err != nil {
		ctx.JSON(fiber.Map{"error": "can not proccess request"})
		return nil
	}

	ctx.JSON(response)
	return nil
}

func (s *SchedulerHandlerImpl) RoundRobin(ctx *fiber.Ctx) error {
	var request *requests.ScheduleRequests
	if err := ctx.BodyParser(request); err != nil {
		ctx.JSON(fiber.Map{
			"error": "invalid request format",
		})
		return nil
	}
	response, err := schedulers.ScheduleRoundRobin(request)
	if err != nil {
		ctx.JSON(fiber.Map{"error": "can not proccess request"})
		return nil
	}

	ctx.JSON(response)
	return nil
}

func (s *SchedulerHandlerImpl) ShortestJobFirst(ctx *fiber.Ctx) error {
	var request *requests.ScheduleRequests
	if err := ctx.BodyParser(request); err != nil {
		ctx.JSON(fiber.Map{
			"error": "invalid request format",
		})
		return nil
	}
	response, err := schedulers.ScheduleShortestJobFirst(request)
	if err != nil {
		ctx.JSON(fiber.Map{"error": "can not proccess request"})
		return nil
	}

	ctx.JSON(response)
	return nil
}

func (s *SchedulerHandlerImpl) MultilevelFeedbackQueue(ctx *fiber.Ctx) error {
	var request *requests.ScheduleRequests
	if err := ctx.BodyParser(request); err != nil {
		ctx.JSON(fiber.Map{
			"error": "invalid request format",
		})
		return nil
	}
	response, err := schedulers.ScheduleMultilevelFeedbackQueue(request)
	if err != nil {
		ctx.JSON(fiber.Map{"error": "can not proccess request"})
		return nil
	}

	ctx.JSON(response)
	return nil
}

func (s *SchedulerHandlerImpl) AllAlgorithms(ctx *fiber.Ctx) error {
	panic("implement me")
}
