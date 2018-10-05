#!/bin/bash -e

cd "$(dirname "${0}")"

go get github.com/taskcluster/generic-worker/cmd/aws-worker-types
"$(go env GOPATH)/bin/aws-worker-types" > /dev/null
cat aws-provisioner-v1-worker-type-definitions/* | sed -n 's/^[[:space:]]*"ImageId": "//p' | sed -n 's/".*//p' | sort -u
rm -r aws-provisioner-v1-worker-type-definitions
