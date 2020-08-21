#!/bin/sh

set -e

# ``readlink -f`` is not our friend on macOS.
SCRIPT_FILE=$(python -c "import os; print(os.path.realpath('${0}'))")
BIN_DIR=$(dirname ${SCRIPT_FILE})
ROOT_DIR=$(dirname ${BIN_DIR})

docker run \
  --rm \
  --volume "${ROOT_DIR}/_bin/generate-tls-certs-on-alpine.sh":/var/code/generate-tls-certs.sh \
  --volume "${ROOT_DIR}/_docker/tls-certs:/var/tls-certs" \
  --env CAROOT=/var/tls-certs \
  golang:1.15.0-alpine3.12 \
  /var/code/generate-tls-certs.sh
