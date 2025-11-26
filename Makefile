# Makefile for Paperless-ngx Merger

.PHONY: help build run clean install test fmt vet lint

# Variabili
BINARY_NAME=paperless-merger
MAIN_PATH=./cmd/paperless-merger
BUILD_DIR=./build
GO=go
VERSION=0.1.4

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the application
	@echo "🔨 Building..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

build-linux: ## Build for Linux AMD64
	@echo "🔨 Building for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "✅ Build completed: $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64"

build-all: ## Build for all platforms (macOS, Linux)
	@echo "🔨 Multi-platform build..."
	@mkdir -p $(BUILD_DIR)
	@echo "  • macOS AMD64..."
	@GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	@echo "  • macOS ARM64..."
	@GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	@echo "  • Linux AMD64..."
	@GOOS=linux GOARCH=amd64 $(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	@echo "✅ Build completed for all platforms"
	@ls -lh $(BUILD_DIR)/

deploy: build install
	@echo "🚀 Application built and installed!"

run: ## Run the application
	@echo "🚀 Starting application..."
	@$(GO) run $(MAIN_PATH)

clean: ## Remove compiled files
	@echo "🧹 Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "✅ Cleaning completed"

install: ## Install the application in GOPATH
	@echo "📦 Installing..."
	@if [ -z "$$GOPATH" ]; then \
		echo "❌ Error: GOPATH is not set"; \
		echo "   Set GOPATH with: export GOPATH=\$$(go env GOPATH)"; \
		exit 1; \
	fi
	@$(GO) install $(MAIN_PATH)
	@echo "✅ Installed: $(BINARY_NAME)"

test: ## Run tests
	@echo "🧪 Running tests..."
	@$(GO) test -v ./...

fmt: ## Format the code
	@echo "✨ Formatting code..."
	@$(GO) fmt ./...

vet: ## Run go vet
	@echo "🔍 Analyzing code..."
	@$(GO) vet ./...

lint: fmt vet ## Run fmt and vet

deps: ## Download dependencies
	@echo "📥 Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "✅ Dependencies updated"

update-deps: ## Update dependencies
	@echo "🔄 Updating dependencies..."
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "✅ Dependencies updated"

release: clean lint test build ## Prepare a release (clean, lint, test, build)
	@echo "🎉 Release ready in $(BUILD_DIR)/"

dev: ## Start in development mode with auto-reload (requires air)
	@which air > /dev/null || (echo "❌ Install air: go install github.com/cosmtrek/air@latest" && exit 1)
	@air

version: ## Show current version
	@echo "$(VERSION)"

bump-patch: ## Increment patch version (e.g. 0.1.0 -> 0.1.1)
	@echo "📦 Updating patch version..."
	@CURRENT_VERSION=$$(echo $(VERSION) | cut -d. -f1-2); \
	PATCH=$$(echo $(VERSION) | cut -d. -f3); \
	NEW_PATCH=$$(($$PATCH + 1)); \
	NEW_VERSION=$$CURRENT_VERSION.$$NEW_PATCH; \
	sed -i '' "s/VERSION=$(VERSION)/VERSION=$$NEW_VERSION/" Makefile; \
	git add Makefile; \
	git commit -m "Bump version to $$NEW_VERSION"; \
	git push; \
	echo "✅ Version updated to $$NEW_VERSION"

bump-minor: ## Increment minor version (e.g. 0.1.0 -> 0.2.0)
	@echo "📦 Updating minor version..."
	@MAJOR=$$(echo $(VERSION) | cut -d. -f1); \
	MINOR=$$(echo $(VERSION) | cut -d. -f2); \
	NEW_MINOR=$$(($$MINOR + 1)); \
	NEW_VERSION=$$MAJOR.$$NEW_MINOR.0; \
	sed -i '' "s/VERSION=$(VERSION)/VERSION=$$NEW_VERSION/" Makefile; \
	git add Makefile; \
	git commit -m "Bump version to $$NEW_VERSION"; \
	git push; \
	echo "✅ Version updated to $$NEW_VERSION"

bump-major: ## Increment major version (e.g. 0.1.0 -> 1.0.0)
	@echo "📦 Updating major version..."
	@MAJOR=$$(echo $(VERSION) | cut -d. -f1); \
	NEW_MAJOR=$$(($$MAJOR + 1)); \
	NEW_VERSION=$$NEW_MAJOR.0.0; \
	sed -i '' "s/VERSION=$(VERSION)/VERSION=$$NEW_VERSION/" Makefile; \
	git add Makefile; \
	git commit -m "Bump version to $$NEW_VERSION"; \
	git push; \
	echo "✅ Version updated to $$NEW_VERSION"

tag: ## Create and push git tag with current version
	@echo "🏷️  Creating tag v$(VERSION)..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@echo "✅ Tag v$(VERSION) created and pushed"

github-release: build-all ## Create a GitHub release with binaries
	@echo "📦 Preparing release v$(VERSION) on GitHub..."
	@rm -rf $(BUILD_DIR)/release
	@mkdir -p $(BUILD_DIR)/release
	@cd $(BUILD_DIR) && \
		tar -czf release/$(BINARY_NAME)-darwin-amd64-v$(VERSION).tar.gz $(BINARY_NAME)-darwin-amd64 && \
		tar -czf release/$(BINARY_NAME)-darwin-arm64-v$(VERSION).tar.gz $(BINARY_NAME)-darwin-arm64 && \
		tar -czf release/$(BINARY_NAME)-linux-amd64-v$(VERSION).tar.gz $(BINARY_NAME)-linux-amd64
	@echo "📝 Generating checksums..."
	@cd $(BUILD_DIR)/release && \
		shasum -a 256 *.tar.gz > checksums.txt
	@echo "🚀 Creating GitHub release..."
	@gh release create v$(VERSION) \
		--title "Release v$(VERSION)" \
		--generate-notes \
		$(BUILD_DIR)/release/*.tar.gz \
		$(BUILD_DIR)/release/checksums.txt
	@echo "✅ Release v$(VERSION) published on GitHub!"

release-patch: bump-patch ## Increment patch, build and publish release
	@$(MAKE) github-release
	@echo "🎉 Patch release completed!"

release-minor: bump-minor ## Increment minor, build and publish release
	@$(MAKE) github-release
	@echo "🎉 Minor release completed!"

release-major: bump-major ## Increment major, build and publish release
	@$(MAKE) github-release
	@echo "🎉 Major release completed!"

.DEFAULT_GOAL := help
