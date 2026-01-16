.PHONY: all linux windows macos run clean install-deps docker-image macos clean-deps

APP_NAME := snap-memory-downloader
BUILD_DIR := bin
SRC_DIR := .
DOCKER_IMAGE := smd-osxcross:darwin

all: linux windows macos

linux:
	@echo "Building for Linux using fyne-cross..."
	fyne-cross linux -arch=amd64 -app-id="com.snapmemory.downloader"
	mkdir -p $(BUILD_DIR)
	mv fyne-cross/bin/linux-amd64/$(APP_NAME) $(BUILD_DIR)/$(APP_NAME) || true
	@echo "Linux build complete: $(BUILD_DIR)/$(APP_NAME)"

windows:
	@echo "Building for Windows using fyne-cross..."
	fyne-cross windows -arch=amd64 -app-id="com.snapmemory.downloader"
	mkdir -p $(BUILD_DIR)
	mv fyne-cross/bin/windows-amd64/$(APP_NAME).exe $(BUILD_DIR)/$(APP_NAME).exe || mv fyne-cross/bin/windows-amd64/$(APP_NAME) $(BUILD_DIR)/$(APP_NAME).exe
	@echo "Windows build complete: $(BUILD_DIR)/$(APP_NAME).exe"

docker-image:
	@echo "Building Docker image: $(DOCKER_IMAGE) (without source code)..."
	docker build -t $(DOCKER_IMAGE) -f Dockerfile .
	@echo "Docker image built: $(DOCKER_IMAGE)"

macos: docker-image
	@echo "Building darwin binary using Docker container..."
	mkdir -p $(BUILD_DIR)
	docker run --rm -v $$(pwd):/root/src $(DOCKER_IMAGE) bash -c "cd /root/src && GOFLAGS=-mod=vendor GOOS=darwin CGO_ENABLED=1 go build -o $(BUILD_DIR)/$(APP_NAME)-darwin -buildvcs=false ./gui_main.go"
	@echo "Darwin build complete: $(BUILD_DIR)/$(APP_NAME)-mac"

run: linux
	@echo "Running application..."
	bin/snap-memory-downloader

clean:
	@echo "Cleaning up..."
	rm -rf $(BUILD_DIR)
	@echo "Cleanup complete."

clean-deps:
	@echo "Removing vendored dependencies..."
	rm -rf vendor/
	@echo "Vendor folder cleaned."

install-deps:
	@echo "Installing dependencies to local vendor folder..."
	go mod download
	go mod vendor
	@echo "Dependencies installed to vendor."