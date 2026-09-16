.PHONY: run build test fmt vet tidy \
        migrate-up migrate-down seed \
        docker-up docker-down docker-logs docker-build \
        postman-validate

APP_ENV ?= .env
MIGRATE_IMAGE := migrate/migrate
# host.docker.internal, not localhost: the migrate container needs to reach
# Postgres running on the host (or port-mapped from another container).
# `--network host` doesn't work on Docker Desktop for Windows/Mac.
DB_URL := postgres://$(or $(DB_USER),postgres):$(or $(DB_PASSWORD),postgres)@$(or $(DB_HOST),host.docker.internal):$(or $(DB_PORT),5432)/$(or $(DB_NAME),football)?sslmode=$(or $(DB_SSLMODE),disable)

## Run the API locally (expects env vars already exported, e.g. via .env)
run:
	go run ./cmd/api

## Build the api and seed binaries into ./bin
build:
	go build -o bin/api ./cmd/api
	go build -o bin/seed ./cmd/seed

## Run all tests
test:
	go test ./... -v

fmt:
	gofmt -l -w .

vet:
	go vet ./...

tidy:
	go mod tidy

## Apply all pending migrations to $(DB_URL). Uses the dockerized
## migrate/migrate image, so no local `migrate` CLI install is required.
## MSYS_NO_PATHCONV avoids Git Bash on Windows mangling the /migrations
## path into a host filesystem path; harmless on Linux/macOS.
migrate-up:
	MSYS_NO_PATHCONV=1 docker run --rm -v "$(CURDIR)/migrations:/migrations" \
		$(MIGRATE_IMAGE) -path=/migrations -database "$(DB_URL)" up

## Roll back all migrations.
migrate-down:
	MSYS_NO_PATHCONV=1 docker run --rm -v "$(CURDIR)/migrations:/migrations" \
		$(MIGRATE_IMAGE) -path=/migrations -database "$(DB_URL)" down -all

## Seed the initial admin user (requires the schema to already be migrated).
## Usage: make seed SEED_USERNAME=admin SEED_PASSWORD=changeme
SEED_USERNAME ?= admin
SEED_PASSWORD ?= admin12345
seed:
	go run ./cmd/seed -username "$(SEED_USERNAME)" -password "$(SEED_PASSWORD)"

## One-command run: app + Postgres + migrations + seed, via docker-compose.
docker-up:
	docker compose up --build -d

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f

docker-build:
	docker compose build

## Check postman_collection.json is well-formed JSON before committing it.
## Not a generator — the collection is hand-maintained; this just catches
## syntax mistakes (e.g. a bad edit) early instead of failing silently on
## import into Postman.
postman-validate:
	python -c "import json; json.load(open('postman_collection.json')); print('postman_collection.json is valid JSON')"
