package handler

import (
	"strconv"

	"github.com/Group-project-bridgeon/purchasing/internal/domain"
	"github.com/Group-project-bridgeon/purchasing/internal/dto"
	"github.com/Group-project-bridgeon/purchasing/internal/service"
	"github.com/gofiber/fiber/v2"
)

type OutboxHandler struct {
	service service.OutboxService
}

func NewOutboxHandler(service service.OutboxService) *OutboxHandler {
	return &OutboxHandler{
		service: service,
	}
}

func (h *OutboxHandler) List(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	var statusPtr *string
	if status := c.Query("status"); status != "" {
		statusPtr = &status
	}

	var aggTypePtr *string
	if aggType := c.Query("aggregate_type"); aggType != "" {
		aggTypePtr = &aggType
	}

	filter := domain.OutboxFilter{
		Status:        statusPtr,
		AggregateType: aggTypePtr,
		Limit:         int32(limit),
		Offset:        int32(offset),
	}

	events, err := h.service.ListEvents(c.UserContext(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	response := dto.ListOutboxEventsResponse{
		Events: dto.ToOutboxEventResponses(events),
		Count:  len(events),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

func (h *OutboxHandler) GetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "event id is required",
		})
	}

	event, err := h.service.GetEventByID(c.UserContext(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "outbox event not found",
		})
	}

	return c.Status(fiber.StatusOK).JSON(dto.ToOutboxEventResponse(event))
}

func (h *OutboxHandler) Retry(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "event id is required",
		})
	}

	if err := h.service.RetryEvent(c.UserContext(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "outbox event queued for retry",
		"id":      id,
	})
}
