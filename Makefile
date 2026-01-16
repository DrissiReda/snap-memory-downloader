.PHONY: all linux windows macos run clean install-deps

APP_NAME := snap-memory-downloader
BUILD_DIR := bin
SRC_DIR := .

all: linux windows macos

linux:
	@echo "Building for Linux..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)/gui_main.go
	@echo "Linux build complete: $(BUILD_DIR)/$(APP_NAME)"

windows:
	@echo "Building for Windows..."
	GOOS=windows go build -ldflags -H=windowsgui -o $(BUILD_DIR)/$(APP_NAME).exe $(SRC_DIR)/gui_main.go
	@echo "Windows build complete: $(BUILD_DIR)/$(APP_NAME).exe"

macos:
	@echo "Building for macOS..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)/gui_main.go
	@echo "macOS build complete: $(BUILD_DIR)/$(APP_NAME)"

run:
	@echo "Running application..."
	go run gui_main.go

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	@echo "Cleanup complete."

install-deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed."