.PHONY: help
help:
	@echo 'Makefile for `golembic` project'
	@echo ''
	@echo 'Usage:'
	@echo '   make dev-deps               Install (or upgrade) development time dependencies'
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
GOLEMBIC_ARGS ?=
GOLEMBIC_SQL_DIR ?= $(shell pwd)/examples/sql

.PHONY: dev-deps
dev-deps:
	go get -v -u github.com/lib/pq
	go get -v -u github.com/spf13/cobra

.PHONY: vet
vet:
	go vet ./...

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
	@PGPASSWORD=$(DB_ADMIN_PASSWORD) \
	  go run ./examples/main.go \
	  --sql-directory $(GOLEMBIC_SQL_DIR) \
	  postgres \
	  --dbname $(DB_NAME) \
	  --host $(DB_HOST) \
	  --port $(DB_PORT) \
	  --ssl-mode $(DB_SSLMODE) \
	  --username $(DB_ADMIN_USER) \
	  $(GOLEMBIC_CMD) $(GOLEMBIC_ARGS)
