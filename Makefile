# Makefile for Simple BlockChain (SBC)

# Variables
APP_NAME_CLI = sbc
APP_NAME_DAEMON = sbcd
BUILD_DIR = bin
CMD_CLI_DIR = ./cmd/sbc
CMD_DAEMON_DIR = ./cmd/sbcd

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get

.PHONY: all build clean test help cli daemon

all: build

build: cli daemon

cli:
	@echo "Building CLI client..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME_CLI) $(CMD_CLI_DIR)

daemon:
	@echo "Building Daemon..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(APP_NAME_DAEMON) $(CMD_DAEMON_DIR)

clean:
	@echo "Cleaning up..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

help:
	@echo "Available commands:"
	@echo "  make build   - Build both CLI and Daemon"
	@echo "  make cli     - Build the CLI client"
	@echo "  make daemon  - Build the Daemon"
	@echo "  make clean   - Remove built binaries"
	@echo "  make test    - Run Go tests"
