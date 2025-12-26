.PHONY: test lint build coverage clean all bench fuzz

# Run all tests with race detection
test:
	go test -v -race ./internal/... ./pkg/...

# Run linter
lint:
	golangci-lint run

# Build the project
build:
	go build ./...

# Generate coverage report
coverage:
	@mkdir -p coverage
	go test -v -coverprofile=coverage/coverage.out ./internal/... ./pkg/...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "Coverage report generated: coverage/coverage.html"
	@go tool cover -func=coverage/coverage.out | grep total

# Clean generated files
clean:
	rm -rf coverage/
	go clean

# Run all checks (test, lint, build, coverage)
all: test lint build coverage

# ================================
# Benchmark Targets
# ================================

# Run all benchmarks with standard settings
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./pkg/xml/

# Run benchmarks and save output to a file
bench-report:
	@mkdir -p benchmarks
	@echo "Running benchmarks and saving to benchmarks/results.txt..."
	go test -bench=. -benchmem ./pkg/xml/ | tee benchmarks/results.txt
	@echo "Benchmark results saved to benchmarks/results.txt"

# ================================
# Fuzz Testing Targets
# ================================

# Run fuzz tests for 30 seconds each
fuzz:
	@echo "Running fuzz tests..."
	@echo "Fuzzing Parse()..."
	go test -fuzz=FuzzParse -fuzztime=30s ./pkg/xml/
	@echo "Fuzzing Validate()..."
	go test -fuzz=FuzzValidate -fuzztime=30s ./pkg/xml/
	@echo "Fuzzing internal parser..."
	go test -fuzz=FuzzParser -fuzztime=30s ./internal/parser/
