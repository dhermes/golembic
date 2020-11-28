#!/bin/sh

set -e

. "$(dirname "${0}")/exists.sh"
. "$(dirname "${0}")/require_env_var.sh"

exists "docker"
requireEnvVar "DB_CONTAINER_NAME"
CONTAINER_EXISTS=$(docker ps --quiet --filter "name=${DB_CONTAINER_NAME}")
if [ -z "${CONTAINER_EXISTS}" ]; then
  echo "Container ${DB_CONTAINER_NAME} is not currently running."
else
  docker rm --force "${DB_CONTAINER_NAME}" > /dev/null
  echo "Container ${DB_CONTAINER_NAME} stopped."
fi

requireEnvVar "DB_NETWORK_NAME"
NETWORK_EXISTS=$(docker network ls --quiet --filter "name=${DB_NETWORK_NAME}")
if [ -z "${NETWORK_EXISTS}" ]; then
  echo "Network ${DB_NETWORK_NAME} is not currently running."
  exit
fi

NETWORK_CONTAINERS=$(docker network inspect --format "{{len .Containers}}" "${DB_NETWORK_NAME}")
if [ "${NETWORK_CONTAINERS}" = "0" ]; then
  docker network rm "${DB_NETWORK_NAME}" > /dev/null
  echo "Network ${DB_NETWORK_NAME} stopped."
else
  echo "Network ${DB_NETWORK_NAME} still has ${NETWORK_CONTAINERS} container(s) running."
fi
