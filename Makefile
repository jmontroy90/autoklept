.PHONY: help test build build-cli build-batch clean

## Show this help
help:
	@awk '/^##/ {sub(/^## /, "", $$0); help=$$0; next} /^[a-zA-Z_-]+:/ {printf "  %-15s %s\n", $$1, help}' $(MAKEFILE_LIST)


## Run all tests
test:
	go mod tidy
	go vet ./...
	go test ./...

## Build all binaries
build: test build-cli build-batch

TARGET_DIR := target
CLI_BIN := $(TARGET_DIR)/autoklept
BATCH_BIN := $(TARGET_DIR)/autoklept-batch

## Ensure target directory exists
$(TARGET_DIR):
	mkdir -p $(TARGET_DIR)

## Build the CLI binary
build-cli: $(TARGET_DIR)
	go build -o $(CLI_BIN) ./cmd/cli

## Build the batch binary
build-batch: $(TARGET_DIR)
	go build -o $(BATCH_BIN) ./cmd/batch

## Remove the target directory
clean:
	rm -rf $(TARGET_DIR)
