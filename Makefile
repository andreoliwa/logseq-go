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

# make -C selects the library directory, including paths with spaces.
LOGSEQ_GO_ROOT := $(CURDIR)
check-upstream: # Report changes since the reviewed upstream snapshot
ifeq ($(OS),Windows_NT)
	@python3 "$(LOGSEQ_GO_ROOT)/scripts/check-upstream.py" || echo warning: upstream check requires python3
else
	@if command -v python3 >/dev/null 2>&1; then python3 "$(LOGSEQ_GO_ROOT)/scripts/check-upstream.py"; else printf '%s\n' 'warning: upstream check requires python3'; fi
endif
.PHONY: check-upstream
