#!/bin/sh

set -e

. "$(dirname "${0}")/exists.sh"
. "$(dirname "${0}")/require_env_var.sh"

requireEnvVar "DB_HOST"
requireEnvVar "DB_PORT"
requireEnvVar "DB_SUPERUSER_NAME"
requireEnvVar "DB_SUPERUSER_USER"
requireEnvVar "DB_SUPERUSER_PASSWORD"
requireEnvVar "DB_NAME"
requireEnvVar "DB_ADMIN_USER"
requireEnvVar "DB_ADMIN_PASSWORD"

DEFAULT_SUPERUSER="root"
DOCKER_BRIDGE_IP="172.17.0.1"

# NOTE: This assumes that `DB_SUPERUSER_USER` / `DB_SUPERUSER_PASSWORD` /
#       `DOCKER_BRIDGE_IP` do not need to be quoted.
#       To verify this rename: `SELECT host, user FROM mysql.user`.
RENAME_ROOT=$(cat <<EOM
CREATE USER ${DB_SUPERUSER_USER}@'${DOCKER_BRIDGE_IP}'
  IDENTIFIED BY '${DB_SUPERUSER_PASSWORD}'
  PASSWORD EXPIRE NEVER;
GRANT ALL PRIVILEGES ON *.* TO
  ${DB_SUPERUSER_USER}@'${DOCKER_BRIDGE_IP}'
WITH GRANT OPTION;
EOM
)

# NOTE: This assumes that `DEFAULT_SUPERUSER` does not need to be quoted.
#       To verify this rename: `SELECT host, user FROM mysql.user`.
DROP_WILDCARD=$(cat <<EOM
DROP USER ${DEFAULT_SUPERUSER}@'%';
DROP USER ${DEFAULT_SUPERUSER}@'localhost';
EOM
)

# NOTE: This assumes that `DB_ADMIN_USER` do not need
#       to be quoted.
CREATE_USER=$(cat <<EOM
CREATE USER ${DB_ADMIN_USER}@'172.17.0.1'
  IDENTIFIED BY '${DB_ADMIN_PASSWORD}'
  PASSWORD EXPIRE NEVER;
EOM
)

# NOTE: This assumes that `DB_NAME` / `DB_ADMIN_USER` do
#       not need to be quoted.
CREATE_DATABASE=$(cat <<EOM
CREATE DATABASE ${DB_NAME};
GRANT ALL PRIVILEGES ON ${DB_NAME}.* TO '${DB_ADMIN_USER}'@'172.17.0.1';
EOM
)

exists "mysql"

echo "${RENAME_ROOT}"
mysql \
  --protocol tcp \
  --database "${DB_SUPERUSER_NAME}" \
  --user "${DEFAULT_SUPERUSER}" \
  --password="${DB_SUPERUSER_PASSWORD}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --execute "${RENAME_ROOT}"

echo "${DROP_WILDCARD}"
mysql \
  --protocol tcp \
  --database "${DB_SUPERUSER_NAME}" \
  --user "${DB_SUPERUSER_USER}" \
  --password="${DB_SUPERUSER_PASSWORD}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --execute "${DROP_WILDCARD}"

echo "${CREATE_USER}"
mysql \
  --protocol tcp \
  --database "${DB_SUPERUSER_NAME}" \
  --user "${DB_SUPERUSER_USER}" \
  --password="${DB_SUPERUSER_PASSWORD}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --execute "${CREATE_USER}"

echo "${CREATE_DATABASE}"
mysql \
  --protocol tcp \
  --database "${DB_SUPERUSER_NAME}" \
  --user "${DB_SUPERUSER_USER}" \
  --password="${DB_SUPERUSER_PASSWORD}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --execute "${CREATE_DATABASE}"
