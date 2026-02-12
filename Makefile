# Go Log Makefile
# Provides utilities for building, testing, and publishing

# Read version from version.txt
VERSION := $(shell cat version.txt)
MODULE := github.com/gopyjs/golog/v3

.PHONY: help build test clean publish-version lint fmt

## help: Show this help message
help:
	@echo "Available commands:"
	@echo "  make build          - Build the project"
	@echo "  make test           - Run all tests"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make publish-version- Publish version to Go proxy"
	@echo "  make tag            - Create git tag for current version"

## build: Build the project
build:
	@echo "Building..."
	go build ./...

## test: Run all tests
test:
	@echo "Running tests..."
	go test -v ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## fmt: Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf ./log
	go clean -cache

## tag: Create git tag for current version
add-tag:
	@echo "Creating git tag v$(VERSION)..."
	git tag v$(VERSION) -m "Release version $(VERSION)"
	git push --tags origin v$(VERSION)
	@echo "Push with: git push origin v$(VERSION)"

del-tag:
	@echo "Deleting git tag v$(VERSION)..."
	git tag -d v$(VERSION)
	git push origin :refs/tags/v$(VERSION)
	@echo "Push with: git push origin :refs/tags/v$(VERSION)"

## publish-version: Publish version to Go proxy (forces go mod cache refresh)
publish-version:
	@echo "Publishing version $(VERSION) to Go proxy..."
	@echo "Triggering Go proxy cache refresh..."
	curl "https://proxy.golang.org/$(MODULE)/@v/v$(VERSION).info" || true
	@echo ""
	@echo "Note: Ensure you have:"
	@echo "  1. Created and pushed git tag: git tag v$(VERSION) && git push origin v$(VERSION)"
	@echo "  2. Module is accessible at: $(MODULE)"
	@echo ""
	@echo "Go proxy will cache the module within a few minutes."
