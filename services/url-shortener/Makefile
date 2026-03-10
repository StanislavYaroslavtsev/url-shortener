include ../../.env
export

export GOOSE_DRIVER := postgres
export GOOSE_MIGRATION_DIR := migrations
export GOOSE_DBSTRING := host=localhost port=$(POSTGRES_PORT) user=$(POSTGRES_USER) password=$(POSTGRES_PASSWORD) dbname=$(POSTGRES_DB) sslmode=$(POSTGRES_SSL_MODE)

migrate-up:
	goose up

migrate-down:
	goose down

migrate-status:
	goose status

lint:
	golangci-lint run