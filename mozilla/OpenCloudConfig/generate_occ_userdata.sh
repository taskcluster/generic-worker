#!/bin/bash -e
THIS_SCRIPT_DIR="$(dirname "${0}")"

# Default directory to look for definitions is current directory.
# To select a different directory, simply export WORKER_TYPES_DIR
# to chosen directory before calling this script.
export WORKER_TYPES_DIR=${WORKER_TYPES_DIR:-.}

go install ./transform-occ

echo "Removing..."
rm -v "${WORKER_TYPES_DIR}"/gecko-*/userdata

echo "Generating..."
curl -L 'https://github.com/mozilla-releng/OpenCloudConfig/tree/master/userdata/Manifest' 2>/dev/null | sed -n 's/.*\(gecko[^.]*\)\.json.*/\1/p' | sort -u | xargs -n 1 -P 32 "${THIS_SCRIPT_DIR}/transform.sh"
