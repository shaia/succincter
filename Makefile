# Go commands
GO := go

# Source files
SRC := ./...

# Targets
.PHONY: all
all: fmt test lint

# Test the library
.PHONY: test
test:
	$(GO) test -v $(SRC)

# Format code
.PHONY: fmt
fmt:
	$(GO) fmt $(SRC)

# Lint (requires golangci-lint installed)
.PHONY: lint
lint:
	golangci-lint run

# Clean up temporary files (no binary cleaning needed for libraries)
.PHONY: clean
clean:
	$(GO) clean

# Generate code (if needed)
.PHONY: generate
generate:
	$(GO) generate $(SRC)

# Install dependencies
.PHONY: deps
deps:
	$(GO) mod tidy