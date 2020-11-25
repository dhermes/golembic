.PHONY: help
help:
	@echo 'Makefile for `golembic` project'
	@echo ''
	@echo 'Usage:'
	@echo '   make dev-deps               Install (or upgrade) development time dependencies'
	@echo '   make vet                    Run `go vet` over source tree'
	@echo '   make shellcheck             Run `shellcheck` on all shell files in `./_bin/`'
	@echo 'PostgreSQL-specific Targets:'
	@echo '   make start-postgres         Starts a PostgreSQL database running in a Docker container and set up users'
	@echo '   make stop-postgres          Stops the PostgreSQL database running in a Docker container'
	@echo '   make restart-postgres       Stops the PostgreSQL database (if running) and starts a fresh Docker container'
	@echo '   make require-postgres       Determine if PostgreSQL database is running; fail if not'
	@echo '   make psql                   Connects to currently running PostgreSQL DB via `psql`'
	@echo '   make run-example-cmd        Run `./examples/cmd/main.go`'
	@echo '   make run-example-script     Run `./examples/script/main.go`'
	@echo 'MySQL-specific Targets:'
	@echo '   make start-mysql            Starts a MySQL database running in a Docker container and set up users'
	@echo '   make stop-mysql             Stops the MySQL database running in a Docker container'
	@echo '   make restart-mysql          Stops the MySQL database (if running) and starts a fresh Docker container'
	@echo '   make require-mysql          Determine if MySQL database is running; fail if not'
	@echo '   make mysql                  Connects to currently running MySQL DB via `mysql`'
	@echo ''

################################################################################
# `make` arguments
################################################################################
ifdef DEBUG
	bindata_flags = -debug
endif

################################################################################
# Meta-variables
################################################################################
SHELLCHECK_PRESENT := $(shell command -v shellcheck 2> /dev/null)

################################################################################
# Environment variable defaults
################################################################################
DB_HOST ?= 127.0.0.1
DB_PORT ?= 18426
DB_SSLMODE ?= disable
DB_CONTAINER_NAME ?= dev-postgres-golembic
DB_NETWORK_NAME ?= dev-network-golembic

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
	go get -v -u github.com/go-sql-driver/mysql
	go get -v -u github.com/lib/pq
	go get -v -u github.com/lib/pq/oid
	go get -v -u github.com/lib/pq/scram
	go get -v -u github.com/spf13/cobra
	go get -v -u github.com/spf13/pflag

.PHONY: vet
vet:
	go vet ./...

.PHONY: _require-shellcheck
_require-shellcheck:
ifndef SHELLCHECK_PRESENT
	$(error 'shellcheck is not installed, it can be installed via "apt-get install shellcheck" or "brew install shellcheck".')
endif

.PHONY: shellcheck
shellcheck: _require-shellcheck
	shellcheck --exclude SC1090 ./_bin/*.sh

################################################################################
# PostgreSQL
################################################################################

.PHONY: start-postgres
start-postgres:
	@DB_NETWORK_NAME=$(DB_NETWORK_NAME) \
	  DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_SUPERUSER_NAME=$(DB_SUPERUSER_NAME) \
	  DB_SUPERUSER_USER=$(DB_SUPERUSER_USER) \
	  DB_SUPERUSER_PASSWORD=$(DB_SUPERUSER_PASSWORD) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/start_postgres.sh

.PHONY: stop-postgres
stop-postgres:
	@DB_NETWORK_NAME=$(DB_NETWORK_NAME) \
	  DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  ./_bin/stop_db.sh

.PHONY: restart-postgres
restart-postgres: stop-postgres start-postgres

.PHONY: require-postgres
require-postgres:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/require_postgres.sh

.PHONY: psql
psql: require-postgres
	@echo "Running psql against port $(DB_PORT)"
	PGPASSWORD=$(DB_ADMIN_PASSWORD) psql \
	  --username $(DB_ADMIN_USER) \
	  --dbname $(DB_NAME) \
	  --port $(DB_PORT) \
	  --host $(DB_HOST)

.PHONY: run-example-cmd
run-example-cmd: require-postgres
	@PGPASSWORD=$(DB_ADMIN_PASSWORD) \
	  go run ./examples/cmd/main.go \
	  --sql-directory $(GOLEMBIC_SQL_DIR) \
	  postgres \
	  --dbname $(DB_NAME) \
	  --host $(DB_HOST) \
	  --port $(DB_PORT) \
	  --ssl-mode $(DB_SSLMODE) \
	  --username $(DB_ADMIN_USER) \
	  $(GOLEMBIC_CMD) $(GOLEMBIC_ARGS)

.PHONY: run-example-script
run-example-script: require-postgres
	@GOLEMBIC_SQL_DIR=$(GOLEMBIC_SQL_DIR) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_NAME=$(DB_NAME) \
	  DB_USER=$(DB_ADMIN_USER) \
	  PGPASSWORD=$(DB_ADMIN_PASSWORD) \
	  DB_SSLMODE=$(DB_SSLMODE) \
	  go run ./examples/script/main.go

################################################################################
# MySQL
################################################################################

.PHONY: start-mysql
start-mysql:
	@DB_NETWORK_NAME=$(DB_NETWORK_NAME) \
	  DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_SUPERUSER_NAME=$(DB_SUPERUSER_NAME) \
	  DB_SUPERUSER_USER=$(DB_SUPERUSER_USER) \
	  DB_SUPERUSER_PASSWORD=$(DB_SUPERUSER_PASSWORD) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/start_mysql.sh

.PHONY: stop-mysql
stop-mysql:
	@DB_NETWORK_NAME=$(DB_NETWORK_NAME) \
	  DB_CONTAINER_NAME=$(DB_CONTAINER_NAME) \
	  ./_bin/stop_db.sh

.PHONY: restart-mysql
restart-mysql: stop-mysql start-mysql

.PHONY: require-mysql
require-mysql:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(DB_PORT) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/require_mysql.sh

.PHONY: mysql
mysql: require-mysql
	@echo "Running mysql against port $(DB_PORT)"
	mysql \
	  --protocol tcp \
	  --user $(DB_ADMIN_USER) \
	  --password=$(DB_ADMIN_PASSWORD) \
	  --database $(DB_NAME) \
	  --port $(DB_PORT) \
	  --host $(DB_HOST)
