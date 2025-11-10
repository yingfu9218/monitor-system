.PHONY: help build build-local build-skip-armv7 release release-snapshot test clean install-goreleaser

# Default target
help:
	@echo "Monitor System - Build Commands"
	@echo ""
	@echo "Available targets:"
	@echo "  make build              - Build using build.sh (local development)"
	@echo "  make build-local        - Build locally using goreleaser (no git tag required)"
	@echo "  make build-skip-armv7   - Build locally skipping ARMv7 (if no cross-compiler)"
	@echo "  make release            - Create release using goreleaser (requires git tag)"
	@echo "  make release-snapshot   - Create snapshot release using goreleaser"
	@echo "  make test               - Run tests"
	@echo "  make clean              - Clean build artifacts"
	@echo "  make install-goreleaser - Install goreleaser"
	@echo ""

# Build using traditional build.sh
build:
	@echo "Building with build.sh..."
	@./build.sh

# Build locally using goreleaser (for testing goreleaser config)
build-local:
	@echo "Building locally with goreleaser..."
	@goreleaser build --snapshot --clean

# Build locally skipping ARMv7 targets (use when cross-compiler is not available)
build-skip-armv7:
	@echo "Building locally with goreleaser (skipping ARMv7)..."
	@goreleaser build --snapshot --clean --skip=server-armv7,agent-armv7

# Create a release (requires a git tag)
release:
	@echo "Creating release with goreleaser..."
	@goreleaser release --clean

# Create a snapshot release (no git tag required)
release-snapshot:
	@echo "Creating snapshot release with goreleaser..."
	@goreleaser release --snapshot --clean

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@rm -rf data/*.db
	@echo "Clean complete!"

# Install goreleaser
install-goreleaser:
	@echo "Installing goreleaser..."
	@go install github.com/goreleaser/goreleaser/v2@latest
	@echo "goreleaser installed successfully!"
	@echo "Verify with: goreleaser --version"
