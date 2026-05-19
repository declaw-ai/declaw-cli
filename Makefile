.PHONY: build build-all clean test install

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GOOS    ?= $(shell go env GOOS)
GOARCH  ?= $(shell go env GOARCH)
BIN_DIR ?= ./bin
BINARY  ?= declaw

LDFLAGS := -s -w \
  -X github.com/declaw-ai/declaw-cli/internal/version.Version=$(VERSION) \
  -X github.com/declaw-ai/declaw-cli/internal/version.Commit=$(COMMIT) \
  -X github.com/declaw-ai/declaw-cli/internal/version.Date=$(DATE)

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -trimpath -ldflags='$(LDFLAGS)' \
		-o $(BIN_DIR)/$(BINARY) ./cmd/declaw/

build-all:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags='$(LDFLAGS)' \
		-o $(BIN_DIR)/$(BINARY)-linux-amd64 ./cmd/declaw/
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags='$(LDFLAGS)' \
		-o $(BIN_DIR)/$(BINARY)-linux-arm64 ./cmd/declaw/
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags='$(LDFLAGS)' \
		-o $(BIN_DIR)/$(BINARY)-darwin-amd64 ./cmd/declaw/
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags='$(LDFLAGS)' \
		-o $(BIN_DIR)/$(BINARY)-darwin-arm64 ./cmd/declaw/
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags='$(LDFLAGS)' \
		-o $(BIN_DIR)/$(BINARY)-windows-amd64.exe ./cmd/declaw/
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags='$(LDFLAGS)' \
		-o $(BIN_DIR)/$(BINARY)-windows-arm64.exe ./cmd/declaw/

test:
	go test -v -race -count=1 ./...

install: build
	cp $(BIN_DIR)/$(BINARY) $(shell go env GOPATH)/bin/$(BINARY)

clean:
	rm -rf $(BIN_DIR)
