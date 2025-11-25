# Makefile per Paperless-ngx Merger

.PHONY: help build run clean install test fmt vet lint

# Variabili
BINARY_NAME=paperless-merger
MAIN_PATH=./cmd/paperless-merger
BUILD_DIR=./build
GO=go
VERSION=0.1.2

help: ## Mostra questo aiuto
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compila l'applicazione
	@echo "🔨 Compilazione in corso..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Compilazione completata: $(BUILD_DIR)/$(BINARY_NAME)"

run: ## Esegue l'applicazione
	@echo "🚀 Avvio applicazione..."
	@$(GO) run $(MAIN_PATH)

clean: ## Rimuove i file compilati
	@echo "🧹 Pulizia in corso..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "✅ Pulizia completata"

install: ## Installa l'applicazione nel GOPATH
	@echo "📦 Installazione in corso..."
	@$(GO) install $(MAIN_PATH)
	@echo "✅ Installato: $(BINARY_NAME)"

test: ## Esegue i test
	@echo "🧪 Esecuzione test..."
	@$(GO) test -v ./...

fmt: ## Formatta il codice
	@echo "✨ Formattazione codice..."
	@$(GO) fmt ./...

vet: ## Esegue go vet
	@echo "🔍 Analisi codice..."
	@$(GO) vet ./...

lint: fmt vet ## Esegue fmt e vet

deps: ## Scarica le dipendenze
	@echo "📥 Download dipendenze..."
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "✅ Dipendenze aggiornate"

update-deps: ## Aggiorna le dipendenze
	@echo "🔄 Aggiornamento dipendenze..."
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "✅ Dipendenze aggiornate"

release: clean lint test build ## Prepara una release (clean, lint, test, build)
	@echo "🎉 Release pronta in $(BUILD_DIR)/"

dev: ## Avvia in modalità sviluppo con auto-reload (richiede air)
	@which air > /dev/null || (echo "❌ Installa air: go install github.com/cosmtrek/air@latest" && exit 1)
	@air

version: ## Mostra la versione corrente
	@echo "$(VERSION)"

bump-patch: ## Incrementa la versione patch (es. 0.1.0 -> 0.1.1)
	@echo "📦 Aggiornamento versione patch..."
	@CURRENT_VERSION=$$(echo $(VERSION) | cut -d. -f1-2); \
	PATCH=$$(echo $(VERSION) | cut -d. -f3); \
	NEW_PATCH=$$(($$PATCH + 1)); \
	NEW_VERSION=$$CURRENT_VERSION.$$NEW_PATCH; \
	sed -i '' "s/VERSION=$(VERSION)/VERSION=$$NEW_VERSION/" Makefile; \
	git add Makefile; \
	git commit -m "Bump version to $$NEW_VERSION"; \
	echo "✅ Versione aggiornata a $$NEW_VERSION"

bump-minor: ## Incrementa la versione minor (es. 0.1.0 -> 0.2.0)
	@echo "📦 Aggiornamento versione minor..."
	@MAJOR=$$(echo $(VERSION) | cut -d. -f1); \
	MINOR=$$(echo $(VERSION) | cut -d. -f2); \
	NEW_MINOR=$$(($$MINOR + 1)); \
	NEW_VERSION=$$MAJOR.$$NEW_MINOR.0; \
	sed -i '' "s/VERSION=$(VERSION)/VERSION=$$NEW_VERSION/" Makefile; \
	git add Makefile; \
	git commit -m "Bump version to $$NEW_VERSION"; \
	echo "✅ Versione aggiornata a $$NEW_VERSION"

bump-major: ## Incrementa la versione major (es. 0.1.0 -> 1.0.0)
	@echo "📦 Aggiornamento versione major..."
	@MAJOR=$$(echo $(VERSION) | cut -d. -f1); \
	NEW_MAJOR=$$(($$MAJOR + 1)); \
	NEW_VERSION=$$NEW_MAJOR.0.0; \
	sed -i '' "s/VERSION=$(VERSION)/VERSION=$$NEW_VERSION/" Makefile; \
	git add Makefile; \
	git commit -m "Bump version to $$NEW_VERSION"; \
	echo "✅ Versione aggiornata a $$NEW_VERSION"

tag: ## Crea e pusha il tag git con la versione corrente
	@echo "🏷️  Creazione tag v$(VERSION)..."
	@git tag -a v$(VERSION) -m "Release v$(VERSION)"
	@git push origin v$(VERSION)
	@echo "✅ Tag v$(VERSION) creato e pushato"

release-patch: bump-patch tag ## Incrementa patch, committa e crea tag
	@echo "🚀 Release patch completata!"

release-minor: bump-minor tag ## Incrementa minor, committa e crea tag
	@echo "🚀 Release minor completata!"

release-major: bump-major tag ## Incrementa major, committa e crea tag
	@echo "🚀 Release major completata!"

.DEFAULT_GOAL := help
