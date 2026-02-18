.PHONY: api

api:
	go run ./cmd/api

migrate:
	go run ./cmd/migrate

build:
	go build -o bin/app ./cmd/api

init:
	docker compose down -v
	docker compose up -d
	sleep 3
	go run ./cmd/migrate
