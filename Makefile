BINARY_NAME := mcp-notion
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS     := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: build test lint run clean fmt tidy

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

test:
	go test -race ./...

cover:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	golangci-lint run ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

tidy:
	go mod tidy

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME) coverage.out
