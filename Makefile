# Exporting bin folder to the path for makefile
export PATH   := $(PWD)/bin:$(PATH)
# Default Shell
export SHELL  := bash
# Type of OS: Linux or Darwin.
export OSTYPE := $(shell uname -s | tr A-Z a-z)
export ARCH := $(shell uname -m)


 
include ./misc/makefile/tools.Makefile

build: test
	@go build ./...

install-deps: gotestsum tparse ## Install Development Dependencies (localy).
deps: $(GOTESTSUM) $(TPARSE) ## Checks for Global Development Dependencies.
deps:
	@echo "Required Tools Are Available"

TESTS_ARGS := --format testname --jsonfile gotestsum.json.out
TESTS_ARGS += --max-fails 2
TESTS_ARGS += -- ./...
TESTS_ARGS += -test.parallel 2
TESTS_ARGS += -test.count    1
TESTS_ARGS += -test.failfast
TESTS_ARGS += -test.coverprofile   coverage.out
TESTS_ARGS += -test.timeout        60s
TESTS_ARGS_WITHRACE := $(TESTS_ARGS)
TESTS_ARGS_WITHRACE += -race

run-tests: $(GOTESTSUM)
	@gotestsum $(TESTS_ARGS_WITHRACE) -short

test: run-tests $(TPARSE) ## Run Tests & parse details
	@cat gotestsum.json.out | $(TPARSE) -all -notests
docker-test:
	@docker-compose -f test.compose.yaml up -d --build 

integration-test: docker-test
	@echo "Running Integration Tests"
	@gotestsum $(TESTS_ARGS)
	@cat gotestsum.json.out | $(TPARSE) -all -notests

integration-test-ci: $(GOTESTSUM) $(TPARSE)
	@echo "Running Integration Tests"
	@gotestsum $(TESTS_ARGS)
	@cat gotestsum.json.out | $(TPARSE) -all -notests

docker-clean:
	@docker-compose -f test.compose.yaml down

lint: $(GOLANGCI) ## Runs golangci-lint with predefined configuration
	@echo "Applying linter"
	golangci-lint version
	golangci-lint run -c .golangci.yaml ./...

.PHONY: lint lint-prepare clean build unittest 


go-generate: $(MOCKERY) ## Runs go generte ./...
	go generate ./...
