#!/bin/sh

set -e

. "$(dirname "${0}")/exists.sh"
. "$(dirname "${0}")/require_env_var.sh"

exists "docker"
requireEnvVar "DB_NETWORK_NAME"
NETWORK_EXISTS=$(docker network ls --quiet --filter "name=${DB_NETWORK_NAME}")
if [ -z "${NETWORK_EXISTS}" ]; then
  docker network create --internal "${DB_NETWORK_NAME}" > /dev/null
  echo "Network ${DB_NETWORK_NAME} created."
else
  echo "Network ${DB_NETWORK_NAME} is already running."
fi

requireEnvVar "DB_CONTAINER_NAME"
STATE_RUNNING=$(docker inspect --format "{{.State.Running}}" "${DB_CONTAINER_NAME}" 2> /dev/null || true)
if [ "${STATE_RUNNING}" = "true" ]; then
  echo "Container ${DB_CONTAINER_NAME} already running."
  exit
fi

STATE_STATUS=$(docker inspect --format "{{.State.Status}}" "${DB_CONTAINER_NAME}" 2> /dev/null || true)
if [ "${STATE_STATUS}" = "exited" ]; then
  echo "Container ${DB_CONTAINER_NAME} is stopped, removing."
  docker rm --force "${DB_CONTAINER_NAME}" > /dev/null
fi

# Get the absolute path to the config file (for Docker)
exists "python"
CONF_DIR="$(dirname "${0}")/../_docker/mysql-conf.d"
# macOS workaround for `readlink`; see https://stackoverflow.com/q/3572030/1068170
CONF_DIR=$(python -c "import os; print(os.path.realpath('${CONF_DIR}'))")

requireEnvVar "DB_HOST"
requireEnvVar "DB_PORT"
requireEnvVar "DB_SUPERUSER_NAME"
requireEnvVar "DB_SUPERUSER_USER"
requireEnvVar "DB_SUPERUSER_PASSWORD"
docker run \
  --detach \
  --hostname "${DB_HOST}" \
  --publish "${DB_PORT}:3306" \
  --name "${DB_CONTAINER_NAME}" \
  --env MYSQL_DATABASE="${DB_SUPERUSER_NAME}" \
  --env MYSQL_USER="${DB_SUPERUSER_USER}" \
  --env MYSQL_PASSWORD="${DB_SUPERUSER_PASSWORD}" \
  --env MYSQL_ROOT_PASSWORD="${DB_SUPERUSER_PASSWORD}" \
  --volume "${CONF_DIR}":/etc/mysql/conf.d \
  mysql:8.0.22 \
  > /dev/null

echo "Container ${DB_CONTAINER_NAME} started on port ${DB_PORT}."

# NOTE: It's crucial to use `docker network connect` vs. starting the container
#       with `docker run --network`. Sine the network is `--internal` any
#       use of `--publish` will be ignored if `--network` is also provided
#       to `docker run`.
docker network connect "${DB_NETWORK_NAME}" "${DB_CONTAINER_NAME}"
echo "Container ${DB_CONTAINER_NAME} added to network ${DB_NETWORK_NAME}."

##########################################################
## Don't exit until `mysqladmin` returns 0 in container ##
##########################################################

# NOTE: This is used strictly for the status code to determine readiness.
mysqladminStatusFull() {
  mysqladmin status \
    --protocol tcp \
    --user "${DB_SUPERUSER_USER}" \
    --password="${DB_SUPERUSER_PASSWORD}" \
    --host "${DB_HOST}" \
    --port "${DB_PORT}"
}

mysqladminStatus() {
  mysqladminStatusFull > /dev/null 2>&1
}

exists "mysqladmin"
# Cap at 300 retries / 30 seconds (by default).
if [ -z "${MYSQLADMIN_STATUS_RETRIES}" ]; then
  MYSQLADMIN_STATUS_RETRIES=300
fi
i=0; while [ ${i} -le ${MYSQLADMIN_STATUS_RETRIES} ]
do
  if mysqladminStatus
  then
    echo "Container ${DB_CONTAINER_NAME} accepting MySQL connections."
    break
  fi
  i=$((i+1))
  sleep "0.1"
done

if [ ${i} -ge ${MYSQLADMIN_STATUS_RETRIES} ]; then
  echo "Container ${DB_CONTAINER_NAME} not accepting MySQL connections."
  echo "  mysqladmin: $(mysqladminStatusFull)"
  exit 1
fi

# Run the superuser migrations
. "$(dirname "${0}")/superuser_migrations_mysql.sh"
