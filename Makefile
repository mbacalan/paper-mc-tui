BINARY  := paper-mc-tui
PKG     := ./cmd/cli
MODULE  := github.com/mbacalan/paper-mc-tui

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X $(MODULE)/internal/buildinfo.Version=$(VERSION) \
	-X $(MODULE)/internal/buildinfo.Commit=$(COMMIT) \
	-X $(MODULE)/internal/buildinfo.Date=$(DATE)

# Platforms built by `make dist` (cross-compiled for manual release uploads).
PLATFORMS := linux/amd64 linux/arm64 windows/amd64 windows/arm64 darwin/amd64 darwin/arm64

.DEFAULT_GOAL := help

## help: list available targets
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed -e 's/## //' | awk -F ': ' '{printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

## build: build the binary for the current platform (version stamped in)
build:
	go build -ldflags '$(LDFLAGS)' -o $(BINARY) $(PKG)

## run: build and run
run:
	go run -ldflags '$(LDFLAGS)' $(PKG)

## test: run tests with the race detector and coverage
test:
	go test -race -cover ./...

## vet: run go vet
vet:
	go vet ./...

## fmt: gofmt -s all Go code in place
fmt:
	gofmt -s -w .

## fmt-check: fail if any file is not gofmt'd
fmt-check:
	@out=$$(gofmt -s -l .); if [ -n "$$out" ]; then echo "not gofmt'd:"; echo "$$out"; exit 1; fi

## tidy: tidy and verify modules
tidy:
	go mod tidy
	go mod verify

## dist: cross-compile release binaries into dist/
dist: clean-dist
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; ext=; [ "$$os" = windows ] && ext=.exe; \
		out=dist/$(BINARY)_$${os}_$${arch}$$ext; \
		echo "  building $$out"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=0 go build -trimpath -ldflags '$(LDFLAGS)' -o $$out $(PKG) || exit 1; \
	done

## clean: remove the binary and dist/
clean: clean-dist
	rm -f $(BINARY)

clean-dist:
	rm -rf dist/

.PHONY: help build run test vet fmt fmt-check tidy dist clean clean-dist
