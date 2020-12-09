#!/bin/sh

set -e

. "$(dirname "${0}")/exists.sh"
. "$(dirname "${0}")/require_env_var.sh"

requireEnvVar "DB_HOST"
requireEnvVar "DB_PORT"
requireEnvVar "DB_ADMIN_DSN"

##########################################################
## Don't exit until `pg_isready` returns 0 in container ##
##########################################################

# NOTE: This is used strictly for the status code to determine readiness.
pgIsReady() {
  pg_isready --dbname "${DB_ADMIN_DSN}" > /dev/null 2>&1
}

exists "pg_isready"
# Cap at 20 retries.
i=0; while [ ${i} -le 20 ]
do
  if pgIsReady
  then
    exit 0
  fi
  i=$((i+1))
  echo "Checking if PostgresSQL is ready on ${DB_HOST}:${DB_PORT} (attempt $i)"
  sleep "0.1"
done

exit 1
