# Makefile per Paperless-ngx Merger

.PHONY: help build run clean install test fmt vet lint

# Variabili
BINARY_NAME=paperless-merger
MAIN_PATH=./cmd/paperless-merger
BUILD_DIR=./build
GO=go

help: ## Mostra questo aiuto
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Compila l'applicazione
	@echo "🔨 Compilazione in corso..."
	@mkdir -p $(BUILD_DIR)
	@$(GO) build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
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

.DEFAULT_GOAL := help
