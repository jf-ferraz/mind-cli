VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
MODULE      = github.com/jf-ferraz/mind-cli
LDFLAGS     = -X $(MODULE)/cmd.Version=$(VERSION) \
              -X $(MODULE)/cmd.CommitSHA=$(COMMIT) \
              -X $(MODULE)/cmd.BuildDate=$(BUILD_DATE)

.PHONY: build install test lint vet clean

## build: compile binary with version info
build:
	go build -ldflags "$(LDFLAGS)" -o mind .

## install: install to $GOPATH/bin as 'mind' with version info
install:
	go build -ldflags "$(LDFLAGS)" -o $$(go env GOPATH)/bin/mind .

## test: run all tests
test:
	go test ./...

## vet: run go vet
vet:
	go vet ./...

## clean: remove build artifacts
clean:
	rm -f mind

## help: show this help
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## //' | column -t -s ':'
