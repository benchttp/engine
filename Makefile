# Default command

.PHONY: default
default:
	@make check

# Check code

.PHONY: check
check:
	@make lint
	@make tests

.PHONY: lint
lint:
	@golangci-lint run

.PHONY: tests
tests:
	@go test -race ./...

TEST_FUNC=^.*$$
ifdef t
TEST_FUNC=$(t)
endif
TEST_PKG=./...
ifdef p
TEST_PKG=./$(p)
endif

.PHONY: test
test:
	@go test -race -timeout 30s -run $(TEST_FUNC) $(TEST_PKG)

# Build

.PHONY: Build
build:
	@go build -v -o ./bin/benchttp ./cmd/runner/main.go

# Docs

.PHONY: docs
docs:
	@echo "\033[4mhttp://localhost:9995/pkg/github.com/benchttp/runner/\033[0m"
	@godoc -http=localhost:9995
