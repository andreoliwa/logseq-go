GO_TEST = go test -v ./... -race -covermode=atomic

help: # Display this help
	@cat Makefile | egrep '^[a-z0-9 ./-]*:.*#' | sed -E -e 's/:.+# */@ /g' -e 's/ .+@/@/g' | sort | awk -F@ '{printf "\033[1;34m%-15s\033[0m %s\n", $$1, $$2}'
.PHONY: help

build: # Build the project
	go mod tidy
	go build
.PHONY: build

test: # Run tests
	$(GO_TEST)
.PHONY: test

test-coverage: # Run tests with coverage
	$(GO_TEST) -coverprofile=coverage-go.out
.PHONY: test-coverage
