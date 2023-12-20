.DEFAULT_GOAL := help

SOURCES := $(shell find . -prune -o -name "*.go" -not -name '*_test.go' -print)

GOIMPORTS ?= goimports

.PHONY: setup
setup:  ## Setup for required tools.
	go get -u golang.org/x/tools/cmd/goimports
	go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	go get -u golang.org/x/tools/cmd/stringer

.PHONY: fmt
fmt: $(SOURCES) ## Formatting source codes.
	@$(GOIMPORTS) -w $^

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.55.2 && \
	golangci-lint run

.PHONY: test
test:
ifdef CI
	go install gotest.tools/gotestsum@v1.8.2
	gotestsum --junitfile test-report.xml -- -v ./...
else
	go test -v ./...
endif

.PHONY: bench
bench:  ## Run benchmarks.
	@go test -bench=. -run=- -benchmem ./...

.PHONY: coverage
cover:  ## Run the tests.
	@go test -coverprofile=coverage.o
	@go tool cover -func=coverage.o

.PHONY: generate
generate: ## Run go generate
	@go generate ./...

.PHONY: build
build: ## Build example command lines.
	./_example/build.sh

.PHONY: help
help: ## Show help text
	@echo "Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2}'
