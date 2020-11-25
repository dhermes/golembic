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

# NOTE: This assumes that `DB_ADMIN_USER` / `DB_SUPERUSER_USER` do not need
#       to be quoted.
CREATE_ROLE=$(cat <<EOM
CREATE ROLE ${DB_ADMIN_USER}
  WITH ENCRYPTED PASSWORD '${DB_ADMIN_PASSWORD}'
  VALID UNTIL 'infinity'
  CONNECTION LIMIT -1
  NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN NOBYPASSRLS NOREPLICATION;
GRANT ${DB_ADMIN_USER} TO ${DB_SUPERUSER_USER};
EOM
)

# NOTE: This assumes that `DB_NAME` / `DB_ADMIN_USER` / `DB_SUPERUSER_USER` do
#       not need to be quoted.
CREATE_DATABASE=$(cat <<EOM
CREATE DATABASE ${DB_NAME}
  OWNER ${DB_ADMIN_USER}
  TEMPLATE "template0"
  ENCODING 'UTF8'
  LC_COLLATE 'en_US.UTF-8'
  LC_CTYPE 'en_US.UTF-8'
  ALLOW_CONNECTIONS true
  CONNECTION LIMIT -1
  IS_TEMPLATE false;
EOM
)

# NOTE: This assumes that `DB_ADMIN_USER` / `DB_SUPERUSER_USER` do not need to
#       be quoted.
REVOKE_ROLE=$(cat <<EOM
REVOKE ${DB_ADMIN_USER} FROM ${DB_SUPERUSER_USER};
EOM
)

exists "psql"

echo "${CREATE_ROLE}"
PGPASSWORD="${DB_SUPERUSER_PASSWORD}" psql \
  --dbname "${DB_SUPERUSER_NAME}" \
  --username "${DB_SUPERUSER_USER}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --command "${CREATE_ROLE}"

echo "${CREATE_DATABASE}"
PGPASSWORD="${DB_SUPERUSER_PASSWORD}" psql \
  --dbname "${DB_SUPERUSER_NAME}" \
  --username "${DB_SUPERUSER_USER}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --command "${CREATE_DATABASE}"

echo "${REVOKE_ROLE}"
PGPASSWORD="${DB_SUPERUSER_PASSWORD}" psql \
  --dbname "${DB_SUPERUSER_NAME}" \
  --username "${DB_SUPERUSER_USER}" \
  --host "${DB_HOST}" \
  --port "${DB_PORT}" \
  --command "${REVOKE_ROLE}"
