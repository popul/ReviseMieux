# Révise mieux - Root Makefile
# Central commands for building, testing, and running the entire project

.PHONY: help install build test lint clean \
        docker-build docker-up docker-down docker-logs docker-clean \
        dev dev-backend dev-frontend db-up db-down db-reset \
        db-migrate db-migrate-down db-migrate-status

# Default target
help:
	@echo "Révise mieux - Available commands:"
	@echo ""
	@echo "  Development:"
	@echo "    make install        Install all dependencies"
	@echo "    make dev            Start full development environment (docker)"
	@echo "    make dev-frontend   Start frontend dev server only"
	@echo "    make dev-backend    Start backend dev server only"
	@echo ""
	@echo "  Build:"
	@echo "    make build          Build all components"
	@echo "    make docker-build   Build all Docker images"
	@echo ""
	@echo "  Testing:"
	@echo "    make test           Run all tests"
	@echo "    make lint           Lint all code"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-up      Start all containers"
	@echo "    make docker-down    Stop all containers"
	@echo "    make docker-logs    View container logs"
	@echo "    make docker-clean   Remove containers and volumes"
	@echo ""
	@echo "  Database:"
	@echo "    make db-up              Start database only"
	@echo "    make db-down            Stop database"
	@echo "    make db-reset           Reset database (delete all data)"
	@echo "    make db-migrate         Apply all pending migrations"
	@echo "    make db-migrate-down    Rollback last migration"
	@echo "    make db-migrate-status  Show migration status"
	@echo ""
	@echo "  Cleanup:"
	@echo "    make clean          Clean build artifacts"

# ============================================================================
# Installation
# ============================================================================

install:
	@echo "Installing backend dependencies..."
	cd backend && go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install
	@echo "Done."

# ============================================================================
# Development
# ============================================================================

# Start full dev environment with Docker Compose
dev: docker-up
	@echo "Development environment started."
	@echo "  Frontend: http://localhost:3000"
	@echo "  Backend:  http://localhost:8080"
	@echo "  Database: localhost:5432"

# Start only the database (for local backend/frontend dev)
db-up:
	docker compose up -d db
	@echo "Database started on localhost:5432"

db-down:
	docker compose stop db

db-reset:
	docker compose down -v db
	docker compose up -d db
	@echo "Database reset complete."

# Run database migrations
db-migrate:
	cd backend && go run ./cmd/migrate -action up

db-migrate-down:
	cd backend && go run ./cmd/migrate -action down

db-migrate-status:
	cd backend && go run ./cmd/migrate -action status

# Run frontend dev server locally (requires npm install first)
dev-frontend:
	cd frontend && npm run dev

# Run backend dev server locally (requires go mod download first)
dev-backend:
	cd backend && go run ./cmd/server

# ============================================================================
# Build
# ============================================================================

build:
	@echo "Building backend..."
	cd backend && $(MAKE) build
	@echo "Building frontend..."
	cd frontend && $(MAKE) build
	@echo "Build complete."

docker-build:
	docker compose build

# ============================================================================
# Testing & Linting
# ============================================================================

test:
	@echo "Testing backend..."
	cd backend && $(MAKE) test
	@echo "Testing frontend..."
	cd frontend && $(MAKE) test
	@echo "All tests passed."

lint:
	@echo "Linting backend..."
	cd backend && $(MAKE) lint
	@echo "Linting frontend..."
	cd frontend && $(MAKE) lint
	@echo "Linting complete."

# ============================================================================
# Docker Compose
# ============================================================================

docker-up:
	docker compose up -d
	@echo "All services started."

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-clean:
	docker compose down -v --remove-orphans
	@echo "Containers and volumes removed."

# ============================================================================
# Cleanup
# ============================================================================

clean:
	@echo "Cleaning backend..."
	cd backend && $(MAKE) clean 2>/dev/null || true
	@echo "Cleaning frontend..."
	cd frontend && $(MAKE) clean 2>/dev/null || rm -rf node_modules dist .next 2>/dev/null || true
	@echo "Clean complete."
