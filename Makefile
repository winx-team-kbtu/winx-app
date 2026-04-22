.PHONY: test test-unit test-coverage test-integration test-all

test: test-unit test-coverage

test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-integration:
	@echo "Running integration tests..."
	go test -v -tags=integration ./...

test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

test-clean:
	@echo "Cleaning test artifacts..."
	rm -f coverage.out coverage.html