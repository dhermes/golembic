.PHONY: help
help:
	@echo 'Makefile for `golembic` project'
	@echo ''
	@echo 'Usage:'
	@echo '   make dev-deps                Install (or upgrade) development time dependencies'
	@echo '   make vet                     Run `go vet` over source tree'
	@echo '   make shellcheck              Run `shellcheck` on all shell files in `./_bin/`'
	@echo 'PostgreSQL-specific Targets:'
	@echo '   make start-postgres          Starts a PostgreSQL database running in a Docker container and set up users'
	@echo '   make stop-postgres           Stops the PostgreSQL database running in a Docker container'
	@echo '   make restart-postgres        Stops the PostgreSQL database (if running) and starts a fresh Docker container'
	@echo '   make require-postgres        Determine if PostgreSQL database is running; fail if not'
	@echo '   make psql                    Connects to currently running PostgreSQL DB via `psql`'
	@echo '   make psql-superuser          Connects to currently running PostgreSQL DB via `psql` as superuser'
	@echo '   make run-postgres-cmd        Run `./examples/cmd/main.go` with `postgres` subcommand'
	@echo '   make run-postgres-example    Run `./examples/postgres-script/main.go`'
	@echo 'MySQL-specific Targets:'
	@echo '   make start-mysql             Starts a MySQL database running in a Docker container and set up users'
	@echo '   make stop-mysql              Stops the MySQL database running in a Docker container'
	@echo '   make restart-mysql           Stops the MySQL database (if running) and starts a fresh Docker container'
	@echo '   make require-mysql           Determine if MySQL database is running; fail if not'
	@echo '   make mysql                   Connects to currently running MySQL DB via `mysql`'
	@echo '   make mysql-superuser         Connects to currently running MySQL DB via `mysql` as superuser'
	@echo '   make run-mysql-cmd           Run `./examples/cmd/main.go` with `mysql` subcommand'
	@echo '   make run-mysql-example       Run `./examples/mysql-script/main.go`'
	@echo 'SQLite-specific Targets:'
	@echo '   make run-sqlite3-example     Run `./examples/sqlite3-script/main.go`'
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
DB_SSLMODE ?= disable
DB_NETWORK_NAME ?= dev-network-golembic

POSTGRES_PORT ?= 18426
POSTGRES_CONTAINER_NAME ?= dev-postgres-golembic
MYSQL_PORT ?= 30892
MYSQL_CONTAINER_NAME ?= dev-mysql-golembic

DB_SUPERUSER_NAME ?= superuser_db
DB_SUPERUSER_USER ?= superuser
DB_SUPERUSER_PASSWORD ?= testpassword_superuser

DB_NAME ?= golembic
DB_ADMIN_USER ?= golembic_admin
DB_ADMIN_PASSWORD ?= testpassword_admin

# NOTE: This assumes the `DB_*_PASSWORD` values do not need to be URL encoded.
POSTGRES_SUPERUSER_DSN ?= postgres://$(DB_SUPERUSER_USER):$(DB_SUPERUSER_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)
POSTGRES_ADMIN_DSN ?= postgres://$(DB_ADMIN_USER):$(DB_ADMIN_PASSWORD)@$(DB_HOST):$(POSTGRES_PORT)/$(DB_NAME)

GOLEMBIC_CMD ?= up
GOLEMBIC_ARGS ?=
GOLEMBIC_SQL_DIR ?= $(shell pwd)/examples/sql

.PHONY: dev-deps
dev-deps:
	go get -v -u github.com/go-sql-driver/mysql
	go get -v -u github.com/lib/pq
	go get -v -u github.com/lib/pq/oid
	go get -v -u github.com/lib/pq/scram
	go get -v -u github.com/mattn/go-sqlite3
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
	  DB_CONTAINER_NAME=$(POSTGRES_CONTAINER_NAME) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(POSTGRES_PORT) \
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
	  DB_CONTAINER_NAME=$(POSTGRES_CONTAINER_NAME) \
	  ./_bin/stop_db.sh

.PHONY: restart-postgres
restart-postgres: stop-postgres start-postgres

.PHONY: require-postgres
require-postgres:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(POSTGRES_PORT) \
	  DB_ADMIN_DSN=$(POSTGRES_ADMIN_DSN) \
	  ./_bin/require_postgres.sh

.PHONY: psql
psql: require-postgres
	@echo "Running psql against port $(POSTGRES_PORT)"
	psql "$(POSTGRES_ADMIN_DSN)"

.PHONY: psql-superuser
psql-superuser: require-postgres
	@echo "Running psql against port $(POSTGRES_PORT)"
	psql "$(POSTGRES_SUPERUSER_DSN)"

.PHONY: run-postgres-cmd
run-postgres-cmd: require-postgres
	@PGPASSWORD=$(DB_ADMIN_PASSWORD) \
	  go run ./examples/cmd/main.go \
	  --sql-directory $(GOLEMBIC_SQL_DIR) \
	  postgres \
	  --dbname $(DB_NAME) \
	  --host $(DB_HOST) \
	  --port $(POSTGRES_PORT) \
	  --ssl-mode $(DB_SSLMODE) \
	  --username $(DB_ADMIN_USER) \
	  $(GOLEMBIC_CMD) $(GOLEMBIC_ARGS)

.PHONY: run-postgres-example
run-postgres-example: require-postgres
	@GOLEMBIC_SQL_DIR=$(GOLEMBIC_SQL_DIR) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(POSTGRES_PORT) \
	  DB_NAME=$(DB_NAME) \
	  DB_USER=$(DB_ADMIN_USER) \
	  PGPASSWORD=$(DB_ADMIN_PASSWORD) \
	  DB_SSLMODE=$(DB_SSLMODE) \
	  go run ./examples/postgres-script/main.go

################################################################################
# MySQL
################################################################################

.PHONY: start-mysql
start-mysql:
	@DB_NETWORK_NAME=$(DB_NETWORK_NAME) \
	  DB_CONTAINER_NAME=$(MYSQL_CONTAINER_NAME) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(MYSQL_PORT) \
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
	  DB_CONTAINER_NAME=$(MYSQL_CONTAINER_NAME) \
	  ./_bin/stop_db.sh

.PHONY: restart-mysql
restart-mysql: stop-mysql start-mysql

.PHONY: require-mysql
require-mysql:
	@DB_HOST=$(DB_HOST) \
	  DB_PORT=$(MYSQL_PORT) \
	  DB_NAME=$(DB_NAME) \
	  DB_ADMIN_USER=$(DB_ADMIN_USER) \
	  DB_ADMIN_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  ./_bin/require_mysql.sh

.PHONY: mysql
mysql: require-mysql
	@echo "Running mysql against port $(MYSQL_PORT)"
	mysql \
	  --protocol tcp \
	  --user $(DB_ADMIN_USER) \
	  --password=$(DB_ADMIN_PASSWORD) \
	  --database $(DB_NAME) \
	  --port $(MYSQL_PORT) \
	  --host $(DB_HOST)

.PHONY: mysql-superuser
mysql-superuser: require-mysql
	@echo "Running mysql against port $(MYSQL_PORT)"
	mysql \
	  --protocol tcp \
	  --user $(DB_SUPERUSER_USER) \
	  --password=$(DB_SUPERUSER_PASSWORD) \
	  --database $(DB_NAME) \
	  --port $(MYSQL_PORT) \
	  --host $(DB_HOST)

.PHONY: run-mysql-cmd
run-mysql-cmd: require-mysql
	@DB_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  go run ./examples/cmd/main.go \
	  --sql-directory $(GOLEMBIC_SQL_DIR) \
	  mysql \
	  --dbname $(DB_NAME) \
	  --host $(DB_HOST) \
	  --port $(MYSQL_PORT) \
	  --user $(DB_ADMIN_USER) \
	  $(GOLEMBIC_CMD) $(GOLEMBIC_ARGS)

.PHONY: run-mysql-example
run-mysql-example: require-mysql
	@GOLEMBIC_SQL_DIR=$(GOLEMBIC_SQL_DIR) \
	  DB_HOST=$(DB_HOST) \
	  DB_PORT=$(MYSQL_PORT) \
	  DB_NAME=$(DB_NAME) \
	  DB_USER=$(DB_ADMIN_USER) \
	  DB_PASSWORD=$(DB_ADMIN_PASSWORD) \
	  go run ./examples/mysql-script/main.go

################################################################################
# SQLite
################################################################################

.PHONY: run-sqlite3-example
run-sqlite3-example:
	@GOLEMBIC_SQL_DIR=$(GOLEMBIC_SQL_DIR) \
	  GOLEMBIC_SQLITE_DB=testing.sqlite3 \
	  go run ./examples/sqlite3-script/main.go
