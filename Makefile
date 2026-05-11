VERSION ?= $(shell git describe --tags --dirty --always 2>/dev/null | sed 's/-[0-9]*-g/-g/' | grep . || echo dev)
LDFLAGS := -X main.version=$(VERSION)
GOLANGCI_LINT_VERSION := v2.11.4
GOBIN := $(or $(shell go env GOBIN),$(shell go env GOPATH)/bin)

.PHONY: all build install test lint lint-gomod clean

all: lint test

build:
	go build -ldflags "$(LDFLAGS)" -o shortcut-statusline .
	cp shortcut-statusline "shortcut-statusline@$(VERSION)"

install:
	go build -ldflags "$(LDFLAGS)" -o "$(GOBIN)/shortcut-statusline" .
	cp "$(GOBIN)/shortcut-statusline" "$(GOBIN)/shortcut-statusline@$(VERSION)"
	@echo "Installed $(GOBIN)/shortcut-statusline and $(GOBIN)/shortcut-statusline@$(VERSION)"

test:
	go test ./...

lint: lint-gomod
	go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION) run ./...

lint-gomod:
	@go mod tidy && \
	if [ -n "$$(git diff --name-only go.mod go.sum)" ]; then \
		echo "error: go.mod is not tidy (go version may have drifted). Run 'go mod tidy' and commit the result." >&2; \
		git checkout go.mod go.sum; \
		exit 1; \
	fi

clean:
	rm -f shortcut-statusline shortcut-statusline@*
