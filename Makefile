.PHONY: help
help:
	@echo 'Makefile for `golembic` project'
	@echo ''
	@echo 'Usage:'
	@echo '   make vet           Run `go vet` over source tree'
	@echo '   make start-db      Starts a PostgreSQL database running in a Docker container'
	@echo '   make stop-db       Stops the PostgreSQL database running in a Docker container'
	@echo '   make require-db    Determine if PostgreSQL database is running; fail if not'
	@echo '   make psql-db       Connects to currently running PostgreSQL DB via `psql`'
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
DB_CONTAINER_NAME ?= dev-postgres-golembic

DB_SUPERUSER_NAME ?= superuser_db
DB_SUPERUSER_USER ?= superuser
DB_SUPERUSER_PASSWORD ?= testpassword_superuser

.PHONY: vet
vet:
	@echo ">> vetting"
	@go vet ./...

.PHONY: start-db
start-db:
	@DB_HOST=$(DB_HOST) \
	  DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  DB_PORT=$(DB_PORT) \
	  DB_SUPERUSER_NAME=$(DB_SUPERUSER_NAME) \
	  DB_SUPERUSER_USER=$(DB_SUPERUSER_USER) \
	  DB_SUPERUSER_PASSWORD=$(DB_SUPERUSER_PASSWORD) \
	  ./_bin/start_db.sh

.PHONY: stop-db
stop-db:
	@DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  ./_bin/stop_db.sh

.PHONY: require-db
require-db:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_SUPERUSER_NAME=$(DB_SUPERUSER_NAME) \
	  DB_SUPERUSER_USER=$(DB_SUPERUSER_USER) \
	  DB_SUPERUSER_PASSWORD=$(DB_SUPERUSER_PASSWORD) \
	  ./_bin/require_db.sh

.PHONY: psql-db
psql-db: require-db
	@echo "Running psql against port $(DB_PORT)"
	PGPASSWORD=$(DB_SUPERUSER_PASSWORD) psql \
	  --username $(DB_SUPERUSER_USER) \
	  --dbname $(DB_SUPERUSER_NAME) \
	  --port $(DB_PORT) \
	  --host $(DB_HOST)
