ifneq (,$(wildcard .env.local))
	include .env.local
else
	include .env
endif
export

GOOSE=goose
MIGRATIONS_DIR=migrations
DB_DRIVER=postgres

.PHONY: migrate-up migrate-down migrate-status migrate-create run build

## Миграции
migrate-up:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DATABASE_URL)" up

migrate-down:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DATABASE_URL)" down

migrate-status:
	$(GOOSE) -dir $(MIGRATIONS_DIR) $(DB_DRIVER) "$(DATABASE_URL)" status

migrate-create:
	@read -p "Migration name: " name; \
	$(GOOSE) -dir $(MIGRATIONS_DIR) create $$name sql

## Docker
docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f postgres

## Приложение
run:
	go run bot/main.go

build:
	go build -o $(BINARY) bot/main.go

## Всё сразу: поднять postgres, накатить миграции, запустить бота
start: docker-up

### TESTS
test:
	go test ./...

test-race:
	go test -race ./...

test-cover:
	go test -cover ./...

test-cover-html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

benchmark:
	go test -bench=. ./...
### TESTS
