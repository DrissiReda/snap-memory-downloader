.PHONY: all linux windows test run clean

APP_NAME := snap-memory-downloader
BUILD_DIR := bin
SRC_DIR := .

all: linux windows

linux:
	@echo "Building for Linux..."
	go build -o $(BUILD_DIR)/$(APP_NAME) $(SRC_DIR)
	@echo "Linux build complete: $(BUILD_DIR)/$(APP_NAME)"

windows:
	@echo "Building for Windows..."
	GOOS=windows go build -o $(BUILD_DIR)/$(APP_NAME).exe $(SRC_DIR)
	@echo "Windows build complete: $(BUILD_DIR)/$(APP_NAME).exe"

test:
	@echo "Running tests..."
	go test ./test/...

run:
	@echo "Running application..."
	go run main.go

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	@echo "Cleanup complete."
