#!/bin/bash -e

############ This script should be used for creating or updating a worker type
############ (i.e. creating AMIs in AWS EC2, and calling the Taskcluster AWS
############ Provisioner API to update the worker type definition with the
############ newly generated AMIs).

# TODO: [pmoore] submit a task after updating worker type

function log {
  echo -e "\x1B[38;5;${COLOUR}m$(date): ${WORKER_TYPE}: ${REGION}: ${@}\x1B[0m"
}

REGION="${1}"
COLOUR="${2}"

if [ -z "${REGION}" ]; then
  echo "Must specify a region to process_region.sh script" >&2
  exit 64
fi

if [ -z "${WORKER_TYPE}" ]; then
  echo "Must export valid WORKER_TYPE env var before calling this script" >&2
  exit 65
fi

if [ -z "${COLOUR}" ]; then
  echo "Missing colour code as input for process_region.sh script" >&2
  exit 66
fi

if [ "${ACTION}" == "update" ]; then
  . "$(dirname "${0}")/update.sh"
elif [ "${ACTION}" == "delete" ]; then
  . "$(dirname "${0}")/delete.sh"
else
  log "$(basename "${0}"): ERROR: Unknown action '${ACTION}' ... exiting" >&2
  exit 86
fi
