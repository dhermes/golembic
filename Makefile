.PHONY: help
help:
	@echo 'Makefile for `golembic` project'
	@echo ''
	@echo 'Usage:'
	@echo '   make vet                    Run `go vet` over source tree'
	@echo '   make start-docker-db        Starts a PostgreSQL database running in a Docker container'
	@echo '   make superuser-migration    Run superuser migration'
	@echo '   make run-migrations         Run all migrations'
	@echo '   make start-db               Run start-docker-db, and migration target(s)'
	@echo '   make stop-db                Stops the PostgreSQL database running in a Docker container'
	@echo '   make restart-db             Stops the PostgreSQL database (if running) and starts a fresh Docker container'
	@echo '   make require-db             Determine if PostgreSQL database is running; fail if not'
	@echo '   make psql-db                Connects to currently running PostgreSQL DB via `psql`'
	@echo '   make run-examples-main      Run `./examples/main.go`'
	@echo ''

################################################################################
# `make` arguments
################################################################################
ifdef DEBUG
	bindata_flags = -debug
endif

################################################################################
# Environment variable defaults
################################################################################
DB_HOST ?= 127.0.0.1
DB_PORT ?= 18426
DB_SSLMODE ?= disable
DB_CONTAINER_NAME ?= dev-postgres-golembic

DB_SUPERUSER_NAME ?= superuser_db
DB_SUPERUSER_USER ?= superuser
DB_SUPERUSER_PASSWORD ?= testpassword_superuser

DB_NAME ?= golembic
DB_ADMIN_USER ?= golembic_admin
DB_ADMIN_PASSWORD ?= testpassword_admin

GOLEMBIC_CMD ?= up
GOLEMBIC_SQL_DIR ?= $(shell pwd)/examples/sql

.PHONY: vet
vet:
	@echo ">> vetting"
	@go vet ./...

.PHONY: start-docker-db
start-docker-db:
	@DB_HOST=$(DB_HOST) \
	  DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  DB_PORT=$(DB_PORT) \
	  DB_SUPERUSER_NAME=$(DB_SUPERUSER_NAME) \
	  DB_SUPERUSER_USER=$(DB_SUPERUSER_USER) \
	  DB_SUPERUSER_PASSWORD=$(DB_SUPERUSER_PASSWORD) \
	  ./_bin/start_db.sh

.PHONY: superuser-migration
superuser-migration:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_SUPERUSER_NAME=$(DB_SUPERUSER_NAME) \
	  DB_SUPERUSER_USER=$(DB_SUPERUSER_USER) \
	  DB_SUPERUSER_PASSWORD=$(DB_SUPERUSER_PASSWORD) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/superuser_migrations.sh

.PHONY: run-migrations
run-migrations: superuser-migration

.PHONY: start-db
start-db: start-docker-db run-migrations

.PHONY: stop-db
stop-db:
	@DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  ./_bin/stop_db.sh

.PHONY: restart-db
restart-db: stop-db start-db

.PHONY: require-db
require-db:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/require_db.sh

.PHONY: psql-db
psql-db: require-db
	@echo "Running psql against port $(DB_PORT)"
	PGPASSWORD=$(DB_ADMIN_PASSWORD) psql \
	  --username $(DB_ADMIN_USER) \
	  --dbname $(DB_NAME) \
	  --port $(DB_PORT) \
	  --host $(DB_HOST)

.PHONY: run-examples-main
run-examples-main: require-db
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_SSLMODE=$(DB_SSLMODE) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  GOLEMBIC_CMD=$(GOLEMBIC_CMD) \
	  GOLEMBIC_SQL_DIR=$(GOLEMBIC_SQL_DIR) \
	  go run ./examples/main.go
