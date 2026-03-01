SHELL := /bin/zsh

APP := diskmon
CMD := ./cmd/diskmon
BUILD_DIR := build

# Override these if your toolchain is installed somewhere else.
# go-duckdb's bundled linux static lib expects glibc, so use GNU toolchains.
LINUX_AMD64_CC ?= x86_64-unknown-linux-gnu-gcc
LINUX_AMD64_CXX ?= x86_64-unknown-linux-gnu-g++
LINUX_ARM64_CC ?= aarch64-unknown-linux-gnu-gcc
LINUX_ARM64_CXX ?= aarch64-unknown-linux-gnu-g++

# Stamp file tracking the last successful webui build.
WEBUI_STAMP := webui/.build-stamp
WEBUI_SOURCES := $(shell find webui/src -type f 2>/dev/null) webui/index.html webui/package.json webui/vite.config.js webui/tailwind.config.js

.PHONY: help deps-check fmt test clean build-webui build-mac build-mac-amd64 build-mac-arm64 build-linux-amd64 build-linux-arm64 build-linux-amd64-nocgo build-linux-arm64-nocgo build-all

help:
	@echo "Targets:"
	@echo "  make build-webui          Build the web UI"
	@echo "  make build-mac            Build native macOS binary for host arch"
	@echo "  make build-mac-amd64      Build macOS amd64 binary with CGO"
	@echo "  make build-mac-arm64      Build macOS arm64 binary with CGO"
	@echo "  make build-linux-amd64    Build Linux amd64 binary with CGO + GNU/glibc cross-compiler"
	@echo "  make build-linux-arm64    Build Linux arm64 binary with CGO + GNU/glibc cross-compiler"
	@echo "  make build-linux-amd64-nocgo Build Linux amd64 binary without CGO (DuckDB disabled at runtime)"
	@echo "  make build-linux-arm64-nocgo Build Linux arm64 binary without CGO (DuckDB disabled at runtime)"
	@echo "  make build-all            Build webui once, then macOS + Linux binaries"
	@echo "  make test                 Run go test ./..."
	@echo "  make fmt                  Run gofmt over all Go files"
	@echo "  make clean                Remove build artifacts"

deps-check:
	@mkdir -p $(BUILD_DIR)

fmt:
	@gofmt -w $(shell rg --files -g '*.go')

test:
	@go test ./...

clean:
	@rm -rf $(BUILD_DIR) $(WEBUI_STAMP)

# Build webui only when source files have changed.
$(WEBUI_STAMP): $(WEBUI_SOURCES)
	@echo "Building WebUI..."
	@cd ./webui && bun run build
	@touch $(WEBUI_STAMP)

build-webui: $(WEBUI_STAMP)

build-mac: deps-check $(WEBUI_STAMP)
	@echo "Building macOS binary for host arch..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=$$(go env GOARCH) go build -o $(BUILD_DIR)/$(APP)-darwin-$$(go env GOARCH) $(CMD)

build-mac-amd64: deps-check $(WEBUI_STAMP)
	@echo "Building macOS amd64 binary..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP)-darwin-amd64 $(CMD)

build-mac-arm64: deps-check $(WEBUI_STAMP)
	@echo "Building macOS arm64 binary..."
	@CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP)-darwin-arm64 $(CMD)

build-linux-amd64: deps-check $(WEBUI_STAMP)
	@command -v $(LINUX_AMD64_CC) >/dev/null || (echo "Missing $(LINUX_AMD64_CC). Install: brew tap messense/macos-cross-toolchains && brew install x86_64-unknown-linux-gnu" && exit 1)
	@command -v $(LINUX_AMD64_CXX) >/dev/null || (echo "Missing $(LINUX_AMD64_CXX). Install: brew tap messense/macos-cross-toolchains && brew install x86_64-unknown-linux-gnu" && exit 1)
	@echo "Building Linux amd64 binary..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=$(LINUX_AMD64_CC) CXX=$(LINUX_AMD64_CXX) go build -o $(BUILD_DIR)/$(APP)-linux-amd64 $(CMD)

build-linux-arm64: deps-check $(WEBUI_STAMP)
	@command -v $(LINUX_ARM64_CC) >/dev/null || (echo "Missing $(LINUX_ARM64_CC). Install: brew tap messense/macos-cross-toolchains && brew install aarch64-unknown-linux-gnu" && exit 1)
	@command -v $(LINUX_ARM64_CXX) >/dev/null || (echo "Missing $(LINUX_ARM64_CXX). Install: brew tap messense/macos-cross-toolchains && brew install aarch64-unknown-linux-gnu" && exit 1)
	@echo "Building Linux arm64 binary..."
	@CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=$(LINUX_ARM64_CC) CXX=$(LINUX_ARM64_CXX) go build -o $(BUILD_DIR)/$(APP)-linux-arm64 $(CMD)

build-linux-amd64-nocgo: deps-check $(WEBUI_STAMP)
	@echo "Building Linux amd64 binary without CGO..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(APP)-linux-amd64-nocgo $(CMD)

build-linux-arm64-nocgo: deps-check $(WEBUI_STAMP)
	@echo "Building Linux arm64 binary without CGO..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(APP)-linux-arm64-nocgo $(CMD)

build-all: $(WEBUI_STAMP) deps-check build-mac build-linux-amd64 build-linux-arm64
	@echo "Built all binaries in $(BUILD_DIR)/"
