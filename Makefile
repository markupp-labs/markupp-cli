GO ?= go
GOLANGCI_LINT ?= $(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
BIN ?= bin/mu
VERSION ?= dev

.PHONY: all test lint fmt vet build cover hooks test-scripts

all: fmt vet lint test build

# Runs every test with the race detector
test:
	$(GO) test -race ./...

# Runs golangci-lint with the repository configuration
lint:
	$(GOLANGCI_LINT) run

# Fails when a Go file is not gofmt formatted, listing the offenders
fmt:
	@out="$$(gofmt -l .)"; if [ -n "$$out" ]; then echo "$$out"; exit 1; fi

vet:
	$(GO) vet ./...

# Builds the static binary without cgo
build:
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags "-s -w -X main.version=$(VERSION)" -o $(BIN) ./cmd/mu

# Writes cover.out and prints the coverage per function
cover:
	$(GO) test ./... -coverprofile=cover.out -covermode=atomic
	$(GO) tool cover -func=cover.out

# Installs the git hooks declared in .pre-commit-config.yaml
hooks:
	pre-commit install --hook-type pre-commit --hook-type commit-msg

# Runs the tests of the shell scripts under scripts/
test-scripts:
	./scripts/valida-mensagem-de-commit-test.sh
