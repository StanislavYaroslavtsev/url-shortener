include ../../.env
export

export GOOSE_DRIVER := clickhouse
export GOOSE_MIGRATION_DIR := migrations
export GOOSE_DBSTRING := clickhouse://$(CLICKHOUSE_USER):$(CLICKHOUSE_PASSWORD)@localhost:9000/$(CLICKHOUSE_DB)

migrate-up:
	goose up

migrate-down:
	goose down

migrate-status:
	goose status

lint:
	golangci-lint run