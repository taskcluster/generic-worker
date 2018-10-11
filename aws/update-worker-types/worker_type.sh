#!/bin/bash -e

############ This script should be used for creating or updating a worker type
############ (i.e. creating AMIs in AWS EC2, and calling the Taskcluster AWS
############ Provisioner API to update the worker type definition with the
############ newly generated AMIs).

# TODO: [pmoore] submit a task after updating worker type

echo "$(date): Checking inputs..."

if [ "${#}" -ne 2 ]; then
  echo "Please provide a worker type and action (delete|update), e.g. worker_type.sh win2012r2 update" >&2
  exit 64
fi

export WORKER_TYPE="${1}"
export ACTION="${2}"

# Default directory to look for definitions is current directory.
# To select a different directory, simply export WORKER_TYPES_DIR
# to chosen directory before calling this script.
WORKER_TYPES_DIR=${WORKER_TYPES_DIR:-.}

# Note we export this env var for subshells, rathing than passing explicitly as
# a command line argument, to keep xargs commands simple later on.
export WORKER_TYPE_DEFINITION_DIR="${WORKER_TYPES_DIR}/${WORKER_TYPE}"

if [ ! -d "${WORKER_TYPE_DEFINITION_DIR}" ]; then
  echo "ERROR: No directory for worker type: '${WORKER_TYPE_DEFINITION_DIR}'" >&2
  echo "Note, if your worker type definitions are stored locally in a different directory, please export WORKER_TYPES_DIR" >&2
  exit 65
fi

echo "$(date): Starting"'!'

# needed to not confuse the script later
rm -f *.latest-ami

# generate a random slugid for aws client token...
go get github.com/taskcluster/slugid-go/slug
go install github.com/taskcluster/generic-worker/aws/cmd/update-worker-type
export SLUGID=$("$(go env GOPATH)/bin/slug")

# aws ec2 describe-regions --query '{A:Regions[*].RegionName}' --output text | grep -v sa-east-1 | while read x REGION; do
# (skip sa-east-1 since it doesn't support all the APIs we use in this script)

echo us-west-1 118 us-west-2 199 us-east-1 100 | xargs -P32 -n2 "$(dirname "${0}")/process_region.sh"

if [ "${ACTION}" == "update" ]; then
  "$(go env GOPATH)/bin/update-worker-type" "${WORKER_TYPE_DEFINITION_DIR}"
  echo
  echo "The worker type has been proactively updated("'!'"):"
  echo
  echo "             https://tools.taskcluster.net/aws-provisioner/#${WORKER_TYPE}/edit"
fi
