#!/bin/sh

set -e

source "$(dirname ${0})/exists.sh"
source "$(dirname ${0})/require_env_var.sh"

exists "docker"
requireEnvVar "DB_CONTAINER_NAME"
EXISTS=$(docker ps --quiet --filter "name=${DB_CONTAINER_NAME}")
if [[ -z "${EXISTS}" ]]; then
    echo "Container ${DB_CONTAINER_NAME} is not currently running."
    exit
fi

docker rm --force "${DB_CONTAINER_NAME}" > /dev/null
echo "Container ${DB_CONTAINER_NAME} stopped."
