#!/bin/sh

requireEnvVar () {
  ENV_VAR="${1}"
  if [ -z "$(eval "echo \$${ENV_VAR}")" ]; then
    echo "${ENV_VAR} environment variable should be set by the caller."
    exit 1
  fi
}
