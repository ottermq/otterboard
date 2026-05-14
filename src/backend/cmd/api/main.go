package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/ottermq/otterboard/src/backend/internal/auth"
	"github.com/ottermq/otterboard/src/backend/internal/config"
	"github.com/ottermq/otterboard/src/backend/internal/db"
	"github.com/ottermq/otterboard/src/backend/internal/routes"
	"github.com/ottermq/otterboard/src/backend/internal/workspaces"
	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := config.LoadConfig()

	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURL)
	defer conn.Close(context.Background())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	queries := db.New(conn)

	app := InitializeFiber(cfg)

	routes.RegisterRoutes(app)

	authService := auth.NewAuthService(queries)
	redisClient := initializeRedis(cfg)
	sessionStore := auth.NewRedisSessionStore(redisClient)
	auth.RegisterAuthRoutes(app, auth.NewHandler(authService, sessionStore, !cfg.DevMode))

	api := routes.RegisterProtectedRoutes(app, auth.AuthMiddleware(sessionStore))

	workspaceService := workspaces.NewWorkspaceService(queries)
	workspaces.RegisterWorkspacesRoutes(api, workspaces.NewHandler(workspaceService))

	log.Fatal(app.Listen(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)))
}

func InitializeFiber(cfg *config.Config) *fiber.App {
	config := fiber.Config{
		Prefork:               !cfg.DevMode,
		ServerHeader:          "Otterboard API",
		AppName:               "Otterboard API",
		DisableStartupMessage: !cfg.DevMode,
	}
	app := fiber.New(config)
	return app
}

func initializeRedis(cfg *config.Config) *redis.Client {
	var numOfAttempts uint = 3
	options, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("failed to parse Redis URL: %v", err)
	}
	client := redis.NewClient(options)
	err = retry.Do(func() error {
		if err := client.Ping(context.Background()).Err(); err != nil {
			return err
		}
		if cfg.DevMode {
			log.Println("Connected to Redis successfully")
		}
		return nil
	},
		retry.Attempts(numOfAttempts),
		retry.Delay(2*time.Second))
	if err != nil {
		log.Fatalf("failed to connect to Redis after %d attempts: %v", numOfAttempts, err)
	}
	return client
}
