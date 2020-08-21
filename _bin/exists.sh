#!/bin/sh

exists() {
  COMMAND="${1}"
  if ! [ -x "$(command -v "${COMMAND}")" ]; then
    echo "Error: ${COMMAND} is not installed." >&2
    exit 1
  fi
}
