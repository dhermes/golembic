#!/bin/sh

set -e -x

# Make sure CAROOT is set (it is used by `mkcert`)
if [[ -z "${CAROOT}" ]]; then
  echo "CAROOT environment variable should be set by the caller."
  exit 1
fi

# NOTE: `git` is needed for `go get`
apk --update --no-cache add git
go get -u github.com/FiloSottile/mkcert

# Clear out any existing root certificate (i.e. we want to always re-generate).
rm -f "${CAROOT}/root-ca-cert.pem" "${CAROOT}/root-ca-key.pem"

# (Re-)generate keys for `localhost`, store them **in** the CA root directory,
# which is expected to be a shared volume with the host.
cd "${CAROOT}"
mkcert --cert-file localhost-cert.pem --key-file localhost-key.pem localhost

# Rename the root CA PEM files (for the benefit of the shared volume on the host).
mv "${CAROOT}/rootCA.pem"     "${CAROOT}/root-ca-cert.pem"
mv "${CAROOT}/rootCA-key.pem" "${CAROOT}/root-ca-key.pem"
