package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ottermq/otterboard/src/backend/internal/config"
	"github.com/ottermq/otterboard/src/backend/internal/routes"
)

func main() {
	cfg := config.LoadConfig()
	app := InitializeFiber(cfg)

	routes.RegisterRoutes(app)

	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))
}
func InitializeFiber(cfg *config.Config) *fiber.App {

	config := fiber.Config{
		Prefork:               !cfg.DevMode, // Disable prefork in development mode for easier debugging
		ServerHeader:          "Otterboard API",
		AppName:               "Otterboard API",
		DisableStartupMessage: !cfg.DevMode, // Disable startup message in production mode
	}
	app := fiber.New(config)
	return app
}
