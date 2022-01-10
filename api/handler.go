package api

import (
	"github.com/gofiber/fiber/v2"
	"os-project/config"
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
	panic("implement me")
}

func (s *SchedulerHandlerImpl) RoundRobin(ctx *fiber.Ctx) error {
	panic("implement me")
}

func (s *SchedulerHandlerImpl) ShortestJobFirst(ctx *fiber.Ctx) error {
	panic("implement me")
}

func (s *SchedulerHandlerImpl) MultilevelFeedbackQueue(ctx *fiber.Ctx) error {
	panic("implement me")
}

func (s *SchedulerHandlerImpl) AllAlgorithms(ctx *fiber.Ctx) error {
	panic("implement me")
}
