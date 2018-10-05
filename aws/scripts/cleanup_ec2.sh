#!/bin/bash -e

cd "$(dirname "${0}")"

go get github.com/taskcluster/generic-worker/cmd/all-worker-types
"$(go env GOPATH)/bin/all-worker-types" > /dev/null
cat worker_type_definitions/* | sed -n 's/^[[:space:]]*"ImageId": "//p' | sed -n 's/".*//p' | sort -u
rm -r worker_type_definitions
