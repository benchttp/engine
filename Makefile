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
	@go test -race -timeout 10s ./...

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
	@go test -race -timeout 10s -run $(TEST_FUNC) $(TEST_PKG)


.PHONY: test-cov
test-cov:
	@go-acc ./...

# Build
.PHONY: build
build:
	@./script/build

.PHONY: clear
clear:
	@rm -rf ./bin/*

# Docs

.PHONY: docs
docs:
	@echo "\033[4mhttp://localhost:9995/pkg/github.com/benchttp/runner/\033[0m"
	@godoc -http=localhost:9995
