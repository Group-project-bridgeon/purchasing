package handler

import (
	"github.com/Group-project-bridgeon/purchasing/internal/middleware"
	"github.com/gofiber/fiber/v2"
)

func clientError(c *fiber.Ctx, err error) string { return middleware.SafeError(c, err) }
