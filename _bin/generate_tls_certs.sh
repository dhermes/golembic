#!/bin/sh

set -e -x

. "$(dirname "${0}")/exists.sh"

exists "python"
# Get the absolute path to root of the repository
ROOT_DIR="$(dirname "${0}")/.."
# macOS workaround for `readlink`; see https://stackoverflow.com/q/3572030/1068170
ROOT_DIR=$(python -c "import os; print(os.path.realpath('${ROOT_DIR}'))")

exists "docker"
docker run \
  --rm \
  --volume "${ROOT_DIR}/_bin":/var/code \
  --volume "${ROOT_DIR}/_docker/tls-certs:/var/tls-certs" \
  --env CAROOT=/var/tls-certs \
  golang:1.15.0-alpine3.12 \
  /var/code/generate_tls_certs_on_alpine.sh
