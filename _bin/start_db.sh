#!/bin/sh

set -e

source "$(dirname ${0})/exists.sh"
source "$(dirname ${0})/require_env_var.sh"

exists "docker"
requireEnvVar "DB_CONTAINER_NAME"
STATUS=$(docker inspect --format "{{.State.Running}}" "${DB_CONTAINER_NAME}" 2> /dev/null || true)
if [[ "${STATUS}" == "true" ]]; then
  echo "Container ${DB_CONTAINER_NAME} already running."
  exit
fi

# Get the absolute path to the config file (for Docker)
exists "python"
CONF_FILE="$(dirname ${0})/../_docker/pg_hba.conf"
# macOS workaround for `readlink`; see https://stackoverflow.com/q/3572030/1068170
CONF_FILE=$(python -c "import os; print(os.path.realpath('${CONF_FILE}'))")

requireEnvVar "DB_HOST"
requireEnvVar "DB_PORT"
requireEnvVar "DB_SUPERUSER_NAME"
requireEnvVar "DB_SUPERUSER_USER"
requireEnvVar "DB_SUPERUSER_PASSWORD"
docker run \
  --detach \
  --hostname "${DB_HOST}" \
  --publish "${DB_PORT}:5432" \
  --name "${DB_CONTAINER_NAME}" \
  --env "POSTGRES_DB=${DB_SUPERUSER_NAME}" \
  --env "POSTGRES_USER=${DB_SUPERUSER_USER}" \
  --env "POSTGRES_PASSWORD=${DB_SUPERUSER_PASSWORD}" \
  --volume "${CONF_FILE}":/etc/postgresql/pg_hba.conf \
  postgres:10.6-alpine \
  -c 'hba_file=/etc/postgresql/pg_hba.conf' \
  > /dev/null

echo "Container ${DB_CONTAINER_NAME} started on port ${DB_PORT}."

##########################################################
## Don't exit until `pg_isready` returns 0 in container ##
##########################################################

# NOTE: This is used strictly for the status code to determine readiness.
pgIsReady() {
  PGPASSWORD="${DB_SUPERUSER_PASSWORD}" pg_isready \
    --dbname "${DB_SUPERUSER_NAME}" \
    --username "${DB_SUPERUSER_USER}" \
    --host "${DB_HOST}" \
    --port "${DB_PORT}" \
    > /dev/null 2>&1
}

exists "pg_isready"
# Cap at 20 retries.
for i in {1..20}
do
  if pgIsReady
  then
    echo "Container ${DB_CONTAINER_NAME} accepting Postgres connections."
    exit 0
  fi
  sleep "0.1"
done

exit 1
