.PHONY: build clean test release dist install help

VERSION ?= $(shell grep 'const version' cmd/bytefall/main.go | cut -d'"' -f2)
LDFLAGS := -s -w -X main.version=$(VERSION)
BINARY := bytefall
DIST_DIR := dist

# Default target
all: build

# Build for current platform
build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/bytefall

# Run tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f $(BINARY)
	rm -rf $(DIST_DIR)

# Install locally
install: build
	sudo cp $(BINARY) /usr/local/bin/

# Build for all platforms
dist: clean
	mkdir -p $(DIST_DIR)
	@echo "Building for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-darwin-arm64 ./cmd/bytefall
	@echo "Building for darwin/amd64..."
	GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-darwin-amd64 ./cmd/bytefall
	@echo "Building for linux/arm64..."
	GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-linux-arm64 ./cmd/bytefall
	@echo "Building for linux/amd64..."
	GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY)-linux-amd64 ./cmd/bytefall
	@echo "Creating archives..."
	cd $(DIST_DIR) && tar -czvf $(BINARY)-darwin-arm64.tar.gz $(BINARY)-darwin-arm64
	cd $(DIST_DIR) && tar -czvf $(BINARY)-darwin-amd64.tar.gz $(BINARY)-darwin-amd64
	cd $(DIST_DIR) && tar -czvf $(BINARY)-linux-arm64.tar.gz $(BINARY)-linux-arm64
	cd $(DIST_DIR) && tar -czvf $(BINARY)-linux-amd64.tar.gz $(BINARY)-linux-amd64
	@echo "Done! Binaries in $(DIST_DIR)/"

# Run in demo mode
demo:
	go run ./cmd/bytefall -demo

# Run with sudo (packet capture)
run: build
	sudo ./$(BINARY)

# Show version
version:
	@echo $(VERSION)

# Generate shell completions
completions:
	@mkdir -p completions
	go run ./cmd/bytefall -completion bash > completions/bytefall.bash
	go run ./cmd/bytefall -completion zsh > completions/_bytefall
	go run ./cmd/bytefall -completion fish > completions/bytefall.fish
	@echo "Completions generated in completions/"

# Show help
help:
	@echo "ByteFall Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build       Build for current platform"
	@echo "  make test        Run tests"
	@echo "  make clean       Remove build artifacts"
	@echo "  make install     Install to /usr/local/bin (requires sudo)"
	@echo "  make dist        Build for all platforms (darwin/linux, arm64/amd64)"
	@echo "  make demo        Run in demo mode"
	@echo "  make run         Build and run with sudo"
	@echo "  make version     Show current version"
	@echo "  make completions Generate shell completion scripts"
	@echo ""
	@echo "Current version: $(VERSION)"
