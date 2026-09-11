.DEFAULT_GOAL := help

setup: ## Setup project for local development
	@test -f .env.local || cp .env.example .env.local

start: ## Start the surge server
	@go run ./cmd/server

build: ## Build the surge server
	@go build -o ./bin/server ./cmd/server

style: ## Runs formatter, static analysis and linter
	@goimports -w .
	@go vet ./...
	@golangci-lint run

tidy: ## Clean module dependencies
	@go mod tidy

clean: ## Remove build artifacts
	@go clean

docker-up: ## Start postgres and redis
	@docker compose up -d

docker-stop: ## Stop containers
	@docker compose stop

docker-down: ## Stop and remove containers
	@docker compose down