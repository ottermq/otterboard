BINARY_NAME=otterboard_api
BACKEND_DIR=src/backend
BUILD_DIR=bin
MAIN_PATH=./cmd/api

build:
	@cd $(BACKEND_DIR) && go mod tidy
	@cd $(BACKEND_DIR) && mkdir -p $(BUILD_DIR)
	@cd $(BACKEND_DIR) && go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

run: build
	@cd $(BACKEND_DIR) && ./$(BUILD_DIR)/$(BINARY_NAME)

test:
	@cd $(BACKEND_DIR) && go test -v ./...

.PHONY: build run test
