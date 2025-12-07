.PHONY: all build run test clean docker-build docker-run migrate-up migrate-down lint

# Variables
BINARY_NAME=main
DOCKER_IMAGE_NAME=cinemaos-backend
GO_FILES=$(shell find . -name '*.go')

all: build

build:
	go build -o tmp/$(BINARY_NAME) ./cmd/api

run: build
	./tmp/$(BINARY_NAME)

dev:
	air -c .air.toml

test:
	go test -v ./...

clean:
	go clean
	rm -f tmp/$(BINARY_NAME)

docker-build:
	docker build -t $(DOCKER_IMAGE_NAME) .

docker-run:
	docker-compose up --build

migrate-up:
	# Add migration command here later
	echo "Running migrations..."

migrate-down:
	# Add rollback command here later
	echo "Rolling back migrations..."

lint:
	golangci-lint run

swag:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal
