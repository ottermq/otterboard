BINARY_NAME=otterboard_api
BACKEND_DIR=src/backend
BUILD_DIR=bin
MIGRATION_DIR=internal/db/migrations
MAIN_PATH=./cmd/api

build:
	@cd $(BACKEND_DIR) && go mod tidy
	@cd $(BACKEND_DIR) && mkdir -p $(BUILD_DIR)
	@cd $(BACKEND_DIR) && go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@cd $(BACKEND_DIR) && ./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@cd $(BACKEND_DIR) && go test -v ./...

migrate-up:
	@cd $(BACKEND_DIR) && migrate -path $(MIGRATION_DIR) -database $(DB_URL) up

migrate-down:
	@cd $(BACKEND_DIR) && migrate -path $(MIGRATION_DIR) -database $(DB_URL) down 1

.PHONY: build run test migrate-up migrate-down
