.PHONY: build run stop down restart clean

# Переменная для именования вашего приложения
APP_NAME=golang-mini-server

# Делает красиво
lint:
	golangci-lint run

# Компиляция и сборка вашего Go-приложения
build:
	@echo "Building the Go application..."
	go build -o $(APP_NAME)

# Запуск контейнеров Docker с использованием Docker Compose
run:
	@echo "Starting Docker containers..."
	docker-compose up --build

# Остановка выполнения контейнеров, но без удаления
stop:
	@echo "Stopping Docker containers..."
	docker-compose stop

# Остановка и удаление всех контейнеров
down:
	@echo "Stopping and removing Docker containers..."
	docker-compose down

# Перезапуск контейнеров
restart: stop run

# Полное очищение: удаление контейнеров, образов, томов
clean:
	@echo "Cleaning up Docker environment..."
	docker-compose down -v --rmi all

# Запуск приложения локально (без контейнеров), если требуется
run-local:
	@echo "Running the Go application locally..."
	./$(APP_NAME)

# Сборка и установка зависимостей для разработки
dev-setup:
	@echo "Setting up development environment..."
	go mod download