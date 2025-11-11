.PHONY: help test test-unit test-corpus test-all lint build clean fmt vet tidy examples ci

LINTER = "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.6.2"

# Default target
help:
	@echo "Available targets:"
	@echo "  make test         - Run unit tests"
	@echo "  make test-corpus  - Run fuzz corpus regression tests"
	@echo "  make test-all     - Run all tests (unit + corpus)"
	@echo "  make lint         - Run golangci-lint"
	@echo "  make fmt          - Format code with gofmt"
	@echo "  make vet          - Run go vet"
	@echo "  make tidy         - Tidy go modules"
	@echo "  make build        - Build the package"
	@echo "  make examples     - Build all examples"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make ci           - Run all CI checks (fmt, vet, lint, test-all)"

# Go environmentHave you 
GOEXPERIMENT ?= jsonv2
GO := GOEXPERIMENT=$(GOEXPERIMENT) go

# Run unit tests
test: test-unit

test-unit:
	$(GO) test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

# Run fuzz corpus regression tests
test-corpus:
	@cd test && $(GO) test -v -run=TestFuzzCorpus || exit 1

# Run all tests
test-all: test-unit test-corpus

# Run linter
lint:
	$(GO) run $(LINTER) run ./... --timeout=5m

# Format code
fmt:
	gofmt -s -w .

# Run go vet
vet:
	$(GO) vet ./...

# Tidy go modules
tidy:
	$(GO) mod tidy
	@cd test && $(GO) mod tidy || exit 1

# Build the package
build:
	$(GO) build ./...

# Build all examples
examples:
	@set -e; \
	for example in basic_usage; do \
		echo "Building example: $$example"; \
		cd examples/$$example && \
		go mod init example 2>/dev/null || true && \
		go mod edit -replace github.com/mikeschinkel/go-rfc9457=../.. && \
		$(GO) mod tidy && \
		$(GO) build -v . || exit 1; \
		cd ../..; \
	done

# Clean build artifacts
clean:
	$(GO) clean
	rm -f coverage.txt

# Run all CI checks locally
ci: fmt vet lint test-all
	@echo "All CI checks passed!"
