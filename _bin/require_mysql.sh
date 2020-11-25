#!/bin/sh

set -e

. "$(dirname "${0}")/exists.sh"
. "$(dirname "${0}")/require_env_var.sh"

requireEnvVar "DB_HOST"
requireEnvVar "DB_PORT"
requireEnvVar "DB_NAME"
requireEnvVar "DB_ADMIN_USER"
requireEnvVar "DB_ADMIN_PASSWORD"

##########################################################
## Don't exit until `mysqladmin` returns 0 in container ##
##########################################################

# NOTE: This is used strictly for the status code to determine readiness.
mysqladminStatus() {
  mysqladmin status \
    --protocol tcp \
    --user "${DB_ADMIN_USER}" \
    --password="${DB_ADMIN_PASSWORD}" \
    --host "${DB_HOST}" \
    --port "${DB_PORT}" \
    > /dev/null 2>&1
}

exists "mysqladmin"
# Cap at 20 retries / 2 seconds (by default).
i=0; while [ ${i} -le 20 ]
do
  if mysqladminStatus
  then
    exit 0
  fi
  i=$((i+1))
  echo "Checking if MySQL is ready on ${DB_HOST}:${DB_PORT} (attempt $i)"
  sleep "0.1"
done

exit 1
