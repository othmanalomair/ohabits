.PHONY: dev build run test clean templ css watch install db-up db-reset help

# Colors for output
GREEN  := \033[0;32m
YELLOW := \033[0;33m
RESET  := \033[0m

## help: Show this help message
help:
	@echo "$(GREEN)ohabits$(RESET) - Personal Habit Tracking App"
	@echo ""
	@echo "Usage:"
	@echo "  make $(YELLOW)<target>$(RESET)"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /'

## dev: Run development server with hot reload
dev:
	@echo "$(GREEN)Starting development server...$(RESET)"
	@make -j3 watch-templ watch-css run-dev

## build: Build production binary
build: templ css
	@echo "$(GREEN)Building production binary...$(RESET)"
	go build -o bin/ohabits cmd/server/main.go

## run: Run the application
run: build
	@echo "$(GREEN)Running application...$(RESET)"
	./bin/ohabits

## run-dev: Run with air for hot reload
run-dev:
	@air -c .air.toml 2>/dev/null || go run cmd/server/main.go

## templ: Generate templ files
templ:
	@echo "$(GREEN)Generating templ files...$(RESET)"
	templ generate

## watch-templ: Watch and regenerate templ files
watch-templ:
	templ generate --watch

## css: Build Tailwind CSS
css:
	@echo "$(GREEN)Building CSS...$(RESET)"
	npx tailwindcss -i ./static/css/input.css -o ./static/css/app.css --minify

## watch-css: Watch and rebuild CSS
watch-css:
	npx tailwindcss -i ./static/css/input.css -o ./static/css/app.css --watch

## install: Install all dependencies
install:
	@echo "$(GREEN)Installing Go dependencies...$(RESET)"
	go mod download
	go mod tidy
	@echo "$(GREEN)Installing Node dependencies...$(RESET)"
	npm install
	@echo "$(GREEN)Installing templ CLI...$(RESET)"
	go install github.com/a-h/templ/cmd/templ@latest
	@echo "$(GREEN)Installing air for hot reload...$(RESET)"
	go install github.com/air-verse/air@latest

## db-up: Create database and run migrations
db-up:
	@echo "$(GREEN)Creating database...$(RESET)"
	psql -U postgres -c "CREATE DATABASE ohabits;" 2>/dev/null || true
	@echo "$(GREEN)Running migrations...$(RESET)"
	psql -d ohabits -f migrations/001_initial.sql

## db-reset: Reset database
db-reset:
	@echo "$(YELLOW)Resetting database...$(RESET)"
	psql -U postgres -c "DROP DATABASE IF EXISTS ohabits;"
	make db-up

## db-shell: Open database shell
db-shell:
	psql ohabits

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning...$(RESET)"
	rm -rf bin/
	rm -rf static/css/app.css
	find templates -name "*_templ.go" -delete

## test: Run tests
test:
	@echo "$(GREEN)Running tests...$(RESET)"
	go test ./...

## lint: Run linter
lint:
	@echo "$(GREEN)Running linter...$(RESET)"
	golangci-lint run

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(RESET)"
	go fmt ./...
	templ fmt templates/
