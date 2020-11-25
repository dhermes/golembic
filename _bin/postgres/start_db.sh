#!/bin/sh

set -e

. "$(dirname "${0}")/../exists.sh"
. "$(dirname "${0}")/../require_env_var.sh"

exists "docker"
requireEnvVar "DB_CONTAINER_NAME"
STATUS=$(docker inspect --format "{{.State.Running}}" "${DB_CONTAINER_NAME}" 2> /dev/null || true)
if [ "${STATUS}" = "true" ]; then
  echo "Container ${DB_CONTAINER_NAME} already running."
  exit
fi

# Get the absolute path to the config file (for Docker)
exists "python"
CONF_FILE="$(dirname "${0}")/../../_docker/pg_hba.conf"
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
pgIsReadyFull() {
  PGPASSWORD="${DB_SUPERUSER_PASSWORD}" pg_isready \
    --dbname "${DB_SUPERUSER_NAME}" \
    --username "${DB_SUPERUSER_USER}" \
    --host "${DB_HOST}" \
    --port "${DB_PORT}"
}

pgIsReady() {
  pgIsReadyFull > /dev/null 2>&1
}

exists "pg_isready"
# Cap at 50 retries / 5 seconds (by default).
if [ -z "${PG_ISREADY_RETRIES}" ]; then
  PG_ISREADY_RETRIES=50
fi
i=0; while [ ${i} -le ${PG_ISREADY_RETRIES} ]
do
  if pgIsReady
  then
    echo "Container ${DB_CONTAINER_NAME} accepting Postgres connections."
    break
  fi
  i=$((i+1))
  sleep "0.1"
done

if [ ${i} -ge ${PG_ISREADY_RETRIES} ]; then
  echo "Container ${DB_CONTAINER_NAME} not accepting Postgres connections."
  echo "  pg_isready: $(pgIsReadyFull)"
  exit 1
fi

# Run the superuser migrations
. "$(dirname "${0}")/superuser_migrations.sh"
