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

# NOTE: This assumes that `DB_ADMIN_USER` do not need
#       to be quoted.
CREATE_USER=$(cat <<EOM
CREATE USER ${DB_ADMIN_USER}
  IDENTIFIED BY '${DB_ADMIN_PASSWORD}'
  PASSWORD EXPIRE NEVER;
EOM
)

# NOTE: This assumes that `DB_NAME` / `DB_ADMIN_USER` do
#       not need to be quoted.
CREATE_DATABASE=$(cat <<EOM
CREATE DATABASE ${DB_NAME};
GRANT ALL PRIVILEGES ON ${DB_NAME}.* TO '${DB_ADMIN_USER}';
EOM
)

exists "mysql"

echo "${CREATE_USER}"
mysql \
  --protocol tcp \
  --database "${DB_SUPERUSER_NAME}" \
  --user root \
  --password="${DB_SUPERUSER_PASSWORD}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --execute "${CREATE_USER}"

echo "${CREATE_DATABASE}"
mysql \
  --protocol tcp \
  --database "${DB_SUPERUSER_NAME}" \
  --user root \
  --password="${DB_SUPERUSER_PASSWORD}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --execute "${CREATE_DATABASE}"
