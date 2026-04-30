package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/ottermq/otterboard/src/backend/internal/auth"
	"github.com/ottermq/otterboard/src/backend/internal/config"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/routes"
)

func main() {
	cfg := config.LoadConfig()

	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	queries := db.New(conn)

	app := InitializeFiber(cfg)

	routes.RegisterRoutes(app)

	authService := auth.NewAuthService(queries)
	auth.RegisterAuthRoutes(app, auth.NewHandler(authService))

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
