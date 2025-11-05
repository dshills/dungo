# Dungeon Generator - Makefile
# Supports: building examples, running tests, quality gates, benchmarks

.PHONY: help build examples test test-short test-verbose test-coverage lint fmt bench clean pre-commit all

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Dungeon Generator - Available Commands"
	@echo ""
	@echo "Building:"
	@echo "  make build          - Build all packages"
	@echo "  make examples       - Build all example programs"
	@echo ""
	@echo "Running Examples:"
	@echo "  make run-zelda      - Generate Zelda-style dungeon (recommended first!)"
	@echo "  make run-metroidvania - Generate Metroidvania-style maze"
	@echo "  make run-darksouls  - Generate Dark Souls challenge dungeon"
	@echo "  make run-all-games  - Run all game-style examples"
	@echo "  make run-text-render - Run text rendering example"
	@echo "  make run-embedding  - Run embedding example"
	@echo ""
	@echo "Testing:"
	@echo "  make test           - Run all tests"
	@echo "  make test-short     - Run tests with -short flag"
	@echo "  make test-verbose   - Run tests with verbose output"
	@echo "  make test-coverage  - Generate test coverage report"
	@echo "  make test-pkg PKG=<name> - Test specific package (e.g., PKG=graph)"
	@echo ""
	@echo "Quality Gates (per Constitution v1.1.1):"
	@echo "  make lint           - Run golangci-lint"
	@echo "  make fmt            - Check code formatting"
	@echo "  make fmt-fix        - Auto-fix formatting issues"
	@echo "  make pre-commit     - Run all pre-commit checks"
	@echo ""
	@echo "Benchmarks:"
	@echo "  make bench          - Run all benchmarks"
	@echo "  make bench-rng      - Benchmark RNG package"
	@echo "  make bench-synthesis - Benchmark synthesis package"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean          - Remove build artifacts"
	@echo "  make deps           - Download dependencies"
	@echo "  make tidy           - Tidy go.mod and go.sum"
	@echo "  make all            - Build everything and run all checks"

## build: Build all packages
build:
	@echo "Building all packages..."
	@go build ./...
	@echo "✓ Build complete"

## examples: Build all example programs
examples:
	@echo "Building examples..."
	@mkdir -p bin
	@go build -o bin/text-render examples/text-render/main.go
	@go build -o bin/embedding examples/embedding/main.go
	@go build -o bin/zelda-dungeon examples/01-zelda-dungeon/main.go
	@go build -o bin/metroidvania-maze examples/02-metroidvania-maze/main.go
	@go build -o bin/darksouls-challenge examples/03-darksouls-challenge/main.go
	@echo "✓ Examples built:"
	@echo "  Technical examples:"
	@echo "    - bin/text-render"
	@echo "    - bin/embedding"
	@echo "  Game-style examples:"
	@echo "    - bin/zelda-dungeon"
	@echo "    - bin/metroidvania-maze"
	@echo "    - bin/darksouls-challenge"

## test: Run all tests
test:
	@echo "Running all tests..."
	@go test ./... -race
	@echo "✓ All tests passed"

## test-short: Run tests with -short flag
test-short:
	@echo "Running short tests..."
	@go test ./... -short
	@echo "✓ Short tests passed"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test ./... -v
	@echo "✓ Verbose tests complete"

## test-coverage: Generate test coverage report
test-coverage:
	@echo "Generating coverage report..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report generated:"
	@echo "  - coverage.out (data)"
	@echo "  - coverage.html (view in browser)"
	@go tool cover -func=coverage.out | grep total | awk '{print "  - Total coverage: " $$3}'

## test-pkg: Test specific package (use: make test-pkg PKG=graph)
test-pkg:
	@if [ -z "$(PKG)" ]; then \
		echo "Error: PKG not specified. Usage: make test-pkg PKG=graph"; \
		exit 1; \
	fi
	@echo "Testing pkg/$(PKG)..."
	@go test ./pkg/$(PKG) -v

## lint: Run golangci-lint (Constitution requirement)
lint:
	@echo "Running golangci-lint..."
	@golangci-lint run
	@echo "✓ Linting passed (zero errors)"

## fmt: Check code formatting (Constitution requirement)
fmt:
	@echo "Checking code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "✗ Files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "✓ All files properly formatted"

## fmt-fix: Auto-fix formatting issues
fmt-fix:
	@echo "Fixing code formatting..."
	@gofmt -w .
	@echo "✓ Formatting applied"

## bench: Run all benchmarks
bench:
	@echo "Running benchmarks..."
	@go test ./... -bench=. -benchmem -run=^$$
	@echo "✓ Benchmarks complete"

## bench-rng: Benchmark RNG package specifically
bench-rng:
	@echo "Benchmarking pkg/rng..."
	@go test ./pkg/rng -bench=. -benchmem -run=^$$

## bench-synthesis: Benchmark synthesis package
bench-synthesis:
	@echo "Benchmarking pkg/synthesis..."
	@go test ./pkg/synthesis -bench=. -benchmem -run=^$$

## pre-commit: Run all pre-commit quality gates (Constitution v1.1.1)
pre-commit:
	@echo "═══════════════════════════════════════════════════════════"
	@echo "Running Pre-Commit Quality Gates (Constitution v1.1.1)"
	@echo "═══════════════════════════════════════════════════════════"
	@echo ""
	@echo "Note: Step 1 (mcp-pr review) must be run manually in Claude Code"
	@echo "      Use: mcp__mcp-pr__review_staged with OpenAI provider"
	@echo ""
	@echo "[2/5] Running golangci-lint..."
	@golangci-lint run
	@echo "✓ Lint passed"
	@echo ""
	@echo "[3/5] Running all tests..."
	@go test ./... -short
	@echo "✓ Tests passed"
	@echo ""
	@echo "[4/5] Checking formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "✗ Files need formatting:"; \
		gofmt -l .; \
		exit 1; \
	fi
	@echo "✓ Formatting check passed"
	@echo ""
	@echo "[5/5] Checking for debug code..."
	@if find pkg -name "*.go" ! -name "*_test.go" -type f -exec grep -l "TODO\|FIXME\|XXX\|HACK" {} + 2>/dev/null; then \
		echo "✗ Found TODO/FIXME markers in production code"; \
		exit 1; \
	fi
	@echo "✓ No debug markers found"
	@echo ""
	@echo "═══════════════════════════════════════════════════════════"
	@echo "✓ All pre-commit checks passed!"
	@echo "Remember: mcp-pr review (step 1) must be done manually"
	@echo "═══════════════════════════════════════════════════════════"

## clean: Remove build artifacts and coverage files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin
	@rm -f coverage.out coverage.html
	@rm -f *.test
	@go clean -cache -testcache
	@echo "✓ Clean complete"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@echo "✓ Dependencies downloaded"

## tidy: Tidy go.mod and go.sum
tidy:
	@echo "Tidying go.mod..."
	@go mod tidy
	@echo "✓ go.mod tidied"

## all: Build everything and run all checks
all: clean deps build examples test lint fmt
	@echo ""
	@echo "═══════════════════════════════════════════════════════════"
	@echo "✓ All builds and checks completed successfully!"
	@echo "═══════════════════════════════════════════════════════════"

# Property-based testing with extended checks
.PHONY: test-property
## test-property: Run property tests with extended checks (1000 iterations)
test-property:
	@echo "Running property-based tests with extended checks..."
	@go test ./... -rapid.checks=1000 -run Property
	@echo "✓ Property tests completed"

# Integration tests
.PHONY: test-integration
## test-integration: Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test ./test/integration/... -v
	@go test ./pkg/dungeon/... -run Integration -v
	@echo "✓ Integration tests complete"

# Golden tests
.PHONY: test-golden
## test-golden: Run golden tests
test-golden:
	@echo "Running golden tests..."
	@go test ./... -run Golden -v

# Update golden test snapshots
.PHONY: update-golden
## update-golden: Update golden test snapshots
update-golden:
	@echo "Updating golden test snapshots..."
	@UPDATE_GOLDEN=1 go test ./... -run Golden -v
	@echo "✓ Golden snapshots updated"

# Run examples
.PHONY: run-text-render run-embedding run-zelda run-metroidvania run-darksouls run-all-games
## run-text-render: Run text rendering example
run-text-render: examples
	@echo "Running text-render example..."
	@./bin/text-render

## run-embedding: Run embedding example
run-embedding: examples
	@echo "Running embedding example..."
	@./bin/embedding

## run-zelda: Run Zelda-style dungeon example
run-zelda:
	@echo "Running Zelda-style dungeon example..."
	@cd examples/01-zelda-dungeon && go run main.go -config config.yaml

## run-metroidvania: Run Metroidvania-style maze example
run-metroidvania:
	@echo "Running Metroidvania-style maze example..."
	@cd examples/02-metroidvania-maze && go run main.go -config config.yaml

## run-darksouls: Run Dark Souls challenge example
run-darksouls:
	@echo "Running Dark Souls challenge example..."
	@cd examples/03-darksouls-challenge && go run main.go -config config.yaml

## run-all-games: Run all game-style examples sequentially
run-all-games:
	@echo "═══════════════════════════════════════════════════════════"
	@echo "Running All Game-Style Examples"
	@echo "═══════════════════════════════════════════════════════════"
	@echo ""
	@echo "[1/3] Zelda-Style Dungeon"
	@echo "───────────────────────────────────────────────────────────"
	@cd examples/01-zelda-dungeon && go run main.go -config config.yaml
	@echo ""
	@echo "[2/3] Metroidvania-Style Maze"
	@echo "───────────────────────────────────────────────────────────"
	@cd examples/02-metroidvania-maze && go run main.go -config config.yaml
	@echo ""
	@echo "[3/3] Dark Souls Challenge"
	@echo "───────────────────────────────────────────────────────────"
	@cd examples/03-darksouls-challenge && go run main.go -config config.yaml
	@echo ""
	@echo "═══════════════════════════════════════════════════════════"
	@echo "✓ All game examples completed!"
	@echo "═══════════════════════════════════════════════════════════"

# Version info
.PHONY: version
## version: Show version information
version:
	@echo "Dungeon Generator"
	@echo "Go version: $$(go version)"
	@echo "Module: $$(go list -m 2>/dev/null || echo 'unknown')"
	@echo "Constitution: v1.1.1"
