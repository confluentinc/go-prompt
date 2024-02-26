### BEGIN MK-INCLUDE UPDATE ###
CURL ?= curl
FIND ?= find
TAR ?= tar

# Mount netrc so curl can work from inside a container
DOCKER_NETRC_MOUNT ?= 1

GITHUB_API = api.github.com
GITHUB_MK_INCLUDE_OWNER := confluentinc
GITHUB_MK_INCLUDE_REPO := cc-mk-include
GITHUB_API_CC_MK_INCLUDE := https://$(GITHUB_API)/repos/$(GITHUB_MK_INCLUDE_OWNER)/$(GITHUB_MK_INCLUDE_REPO)
GITHUB_API_CC_MK_INCLUDE_TARBALL := $(GITHUB_API_CC_MK_INCLUDE)/tarball
GITHUB_API_CC_MK_INCLUDE_VERSION ?= $(GITHUB_API_CC_MK_INCLUDE_TARBALL)/$(MK_INCLUDE_VERSION)

MK_INCLUDE_DIR := mk-include
MK_INCLUDE_LOCKFILE := .mk-include-lockfile
MK_INCLUDE_TIMESTAMP_FILE := .mk-include-timestamp
# For optimum performance, you should override MK_INCLUDE_TIMEOUT_MINS above the managed section headers to be
# a little longer than the worst case cold build time for this repo.
MK_INCLUDE_TIMEOUT_MINS ?= 240
# If this latest validated release is breaking you, please file a ticket with DevProd describing the issue, and
# if necessary you can temporarily override MK_INCLUDE_VERSION above the managed section headers until the bad
# release is yanked.
MK_INCLUDE_VERSION ?= v0.1012.0

# Make sure we always have a copy of the latest cc-mk-include release less than $(MK_INCLUDE_TIMEOUT_MINS) old:
./$(MK_INCLUDE_DIR)/%.mk: .mk-include-check-FORCE
	@trap "rm -f $(MK_INCLUDE_LOCKFILE); exit" 0 2 3 15; \
	waitlock=0; while ! ( set -o noclobber; echo > $(MK_INCLUDE_LOCKFILE) ); do \
	   sleep $$waitlock; waitlock=`expr $$waitlock + 1`; \
	   test 14 -lt $$waitlock && { \
	      echo 'stealing stale lock after 105s' >&2; \
	      break; \
	   } \
	done; \
	test -s $(MK_INCLUDE_TIMESTAMP_FILE) || rm -f $(MK_INCLUDE_TIMESTAMP_FILE); \
	test -z "`$(FIND) $(MK_INCLUDE_TIMESTAMP_FILE) -mmin +$(MK_INCLUDE_TIMEOUT_MINS) 2>&1`" || { \
	   grep -q 'machine $(GITHUB_API)' ~/.netrc 2>/dev/null || { \
	      echo 'error: follow https://confluentinc.atlassian.net/l/cp/0WXXRLDh to fix your ~/.netrc'; \
	      exit 1; \
	   }; \
	   $(CURL) --fail --silent --netrc --location "$(GITHUB_API_CC_MK_INCLUDE_VERSION)" --output $(MK_INCLUDE_TIMESTAMP_FILE)T --write-out '$(GITHUB_API_CC_MK_INCLUDE_VERSION): %{errormsg}\n' >&2 \
	   && $(TAR) zxf $(MK_INCLUDE_TIMESTAMP_FILE)T \
	   && rm -rf $(MK_INCLUDE_DIR) \
	   && mv $(GITHUB_MK_INCLUDE_OWNER)-$(GITHUB_MK_INCLUDE_REPO)-* $(MK_INCLUDE_DIR) \
	   && mv -f $(MK_INCLUDE_TIMESTAMP_FILE)T $(MK_INCLUDE_TIMESTAMP_FILE) \
	   && echo 'installed cc-mk-include-$(MK_INCLUDE_VERSION) from $(GITHUB_MK_INCLUDE_REPO)' \
	   ; \
	} || { \
	   rm -f $(MK_INCLUDE_TIMESTAMP_FILE)T; \
	   if test -f $(MK_INCLUDE_TIMESTAMP_FILE); then \
	      touch $(MK_INCLUDE_TIMESTAMP_FILE); \
	      echo 'unable to access $(GITHUB_MK_INCLUDE_REPO) fetch API to check for latest release; next try in $(MK_INCLUDE_TIMEOUT_MINS) minutes' >&2; \
	   else \
	      echo 'unable to access $(GITHUB_MK_INCLUDE_REPO) fetch API to bootstrap mk-include subdirectory' >&2 && false; \
	   fi; \
	}

.PHONY: .mk-include-check-FORCE
.mk-include-check-FORCE:
	@test -z "`git ls-files $(MK_INCLUDE_DIR)`" || { \
		echo 'fatal: checked in $(MK_INCLUDE_DIR)/ directory is preventing make from fetching recent cc-mk-include releases for CI' >&2; \
		exit 1; \
	}
### END MK-INCLUDE UPDATE ###
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
