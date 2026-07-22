-include .env
export
BINARY=go-social-feed

infra-up:
	@echo "Starting MongoDB and RabbitMQ"
	docker compose -f deployment/docker-compose.yml up -d mongodb rabbitmq

rabbitmq-up:
	@echo "Starting RabbitMQ"
	docker compose -f deployment/docker-compose.yml up -d rabbitmq

infra-down:
	docker compose -f deployment/docker-compose.yml down

run-api: rabbitmq-up
	@echo "Running the application"
	go run cmd/api/main.go

swagger:
	@echo "Generating swagger documentation..."
	swag init -g cmd/api/main.go -o docs
