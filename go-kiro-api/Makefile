.PHONY: build run clean test docker-build docker-run install

# Build the application
build:
	go build -o kiro-api cmd/server/main.go

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o kiro-api-linux-amd64 cmd/server/main.go
	GOOS=windows GOARCH=amd64 go build -o kiro-api-windows-amd64.exe cmd/server/main.go
	GOOS=darwin GOARCH=amd64 go build -o kiro-api-darwin-amd64 cmd/server/main.go
	GOOS=darwin GOARCH=arm64 go build -o kiro-api-darwin-arm64 cmd/server/main.go

# Run the application
run:
	go run cmd/server/main.go

# Clean build artifacts
clean:
	rm -f kiro-api kiro-api-*

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Install dependencies
install:
	go mod download
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Build Docker image
docker-build:
	docker build -t kiro-api:latest .

# Run Docker container
docker-run:
	docker run -p 3000:3000 \
		-e KIRO_CREDS_BASE64=${KIRO_CREDS_BASE64} \
		-e KIRO_REGION=${KIRO_REGION} \
		kiro-api:latest

# Development mode with auto-reload (requires air)
dev:
	air
