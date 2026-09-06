PROJECT := thumbpress
BINARY := bin/$(PROJECT)
INSTALL_DIR ?= $(HOME)/.local/bin

.PHONY: help
help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build bin/thumbpress
	@mkdir -p bin
	go build -o $(BINARY) ./cmd/$(PROJECT)

.PHONY: run
run: ## Run thumbpress with ARGS="..."
	go run ./cmd/$(PROJECT) $(ARGS)

.PHONY: test
test: ## Run the test suite
	go test -v ./...

.PHONY: lint
lint: ## Run static analysis
	go vet ./...

.PHONY: fmt
fmt: ## Format Go source files
	go fmt ./...

.PHONY: fix
fix: ## Apply Go modernization fixes
	go fix ./...

.PHONY: tidy
tidy: ## Synchronize module dependencies
	go mod tidy

.PHONY: install
install: ## Install thumbpress to INSTALL_DIR (default ~/.local/bin)
	@mkdir -p $(INSTALL_DIR) && go build -o $(INSTALL_DIR)/$(PROJECT) ./cmd/$(PROJECT)

.PHONY: clean
clean: ## Remove generated binaries
	rm -rf bin/
