APP_NAME := mux-ssh
BUILD_DIR := build
CMD_PATH := ./cmd/ssh-ogm/main.go

.PHONY: all build build-all clean

all: build

build:
	go build -o $(APP_NAME) $(CMD_PATH)

build-all: clean
	mkdir -p $(BUILD_DIR)
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 $(CMD_PATH)
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 $(CMD_PATH)
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe $(CMD_PATH)
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 $(CMD_PATH)
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 $(CMD_PATH)
	@echo "Done! Binaries are in $(BUILD_DIR)/"

clean:
	rm -rf $(BUILD_DIR)
	rm -f $(APP_NAME)
