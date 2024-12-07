.PHONY: build run stop down restart clean

APP_NAME=golang-mini-server

# Docker
run: down build up

build:
	@echo "Building Docker containers..."
	docker-compose build

up:
	@echo "Starting Docker containers..."
	docker-compose up

stop:
	@echo "Stopping Docker containers..."
	docker-compose stop

down:
	@echo "Stopping and removing Docker containers..."
	docker-compose down

restart: stop run

clean:
	@echo "Cleaning up Docker environment..."
	docker-compose down -v --rmi all

# Golang
lint:
	golangci-lint run

dev-build:
	@echo "Building the Go application..."
	go build -o $(APP_NAME)

run-local:
	@echo "Running the Go application locally..."
	./$(APP_NAME)

dev-setup:
	@echo "Setting up development environment..."
	go mod download
