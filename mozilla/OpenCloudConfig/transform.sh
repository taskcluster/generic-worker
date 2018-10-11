#!/bin/bash -e

echo "${1}"
mkdir -p "${WORKER_TYPES_DIR}/${1}"
"$(go env GOPATH)/bin/transform-occ" "${1}" > "${WORKER_TYPES_DIR}/${1}/userdata" || rm "${WORKER_TYPES_DIR}/${1}/userdata"
