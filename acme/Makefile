
# Variables
BINARY_NAME=acme
BUILD_DIR=build
MAIN_PATH=cmd/main.go

# Build the application
build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Run the application
run:
	go run $(MAIN_PATH)

# Test the application
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -rf $(BUILD_DIR)

# Install dependencies
install-deps:
	go mod tidy
	go mod download

# Build for multiple platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

# Docker build
podman-build:
	podman build -t $(BINARY_NAME):latest .

# Run with Docker
podman-run:
	podman run --rm -v ~/.kube:/root/.kube $(BINARY_NAME):latest