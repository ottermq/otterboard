package stats

import "github.com/gofiber/fiber/v2"

func RegisterStatsRoutes(wsGroup fiber.Router, h *Handler) {
	wsGroup.Get("/stats", h.GetStats)
}
