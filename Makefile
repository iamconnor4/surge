.DEFAULT_GOAL := help

setup: ## Setup project for local development
	@test -f .env.local || cp .env.example .env.local

start: ## Start the surge server
	@go run ./cmd/server

build: ## Build the surge server
	@go build -o ./bin/server ./cmd/server

check: ## Run CI checks
	@go vet ./...
	@go tool golangci-lint run
	@go test -race ./...
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum

tidy: ## Format source and clean module dependencies
	@go tool goimports -w .
	@go mod tidy

clean: ## Remove build artifacts
	@go clean

test: ## Run tests
	@go test ./...

docker-up: ## Start postgres and redis
	@docker compose up -d

docker-stop: ## Stop containers
	@docker compose stop

docker-down: ## Stop and remove containers
	@docker compose down

migration: ## Create a migration file
	@if test -z "$(title)"; then \
		echo "Usage: make migration title=some_title"; \
	else \
		goose \
			-dir ./internal/platform/postgres/migrations \
			create \
			$(title) \
			sql; \
	fi

db-up: ## Migrate the DB to the most recent version
	@goose \
		-dir ./internal/platform/postgres/migrations \
		postgres \
  		"postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}" \
  		up

db-down: ## Rollback the DB version by 1
	@goose \
		-dir ./internal/platform/postgres/migrations \
		postgres \
  		"postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}" \
  		down

db-status: ## Dump migration status for the current DB
	@goose \
		-dir ./internal/platform/postgres/migrations \
		postgres \
  		"postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=${POSTGRES_SSLMODE}" \
  		status
