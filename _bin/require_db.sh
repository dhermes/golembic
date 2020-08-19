#!/bin/sh

set -e

source "$(dirname ${0})/exists.sh"
source "$(dirname ${0})/require_env_var.sh"

requireEnvVar "DB_HOST"
requireEnvVar "DB_PORT"
requireEnvVar "DB_NAME"
requireEnvVar "DB_ADMIN_USER"
requireEnvVar "DB_ADMIN_PASSWORD"

##########################################################
## Don't exit until `pg_isready` returns 0 in container ##
##########################################################

# NOTE: This is used strictly for the status code to determine readiness.
pgIsReady() {
  PGPASSWORD="${DB_ADMIN_PASSWORD}" pg_isready \
    --dbname "${DB_NAME}" \
    --username "${DB_ADMIN_USER}" \
    --host "${DB_HOST}" \
    --port "${DB_PORT}" \
    > /dev/null 2>&1
}

exists "pg_isready"
# TODO: https://github.com/dhermes/golembic/issues/3
until pgIsReady
do
  echo "Checking if PostgresSQL is ready on ${DB_HOST}:${DB_PORT}"
  sleep "0.1"
done
