.PHONY: build test lint clean install

VERSION ?= dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X github.com/sysrqio/zuginbox/internal/version.Version=$(VERSION) \
	-X github.com/sysrqio/zuginbox/internal/version.Commit=$(COMMIT) \
	-X github.com/sysrqio/zuginbox/internal/version.Date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o bin/zuginbox ./cmd/zuginbox

test:
	go test ./...

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/zuginbox

clean:
	rm -rf bin/

lint:
	go vet ./...
	go fmt ./...
