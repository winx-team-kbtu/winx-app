.PHONY: test test-coverage test-race test-all clean

test:
	@echo "Running all tests..."
	go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

test-verbose:
	@echo "Running tests in verbose mode..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

test-services:
	@echo "Testing services..."
	go test -v ./internal/app/domain/services/...

test-handlers:
	@echo "Testing handlers..."
	go test -v ./internal/app/domain/handlers/...

test-store:
	@echo "Testing notification store..."
	go test -v ./internal/app/notifications/...

clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html
	go clean -testcache

test-all: clean test test-race test-coverage
	@echo "All tests completed!"