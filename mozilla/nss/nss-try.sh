#!/bin/bash -xve

######
# This script tests the *latest* generic-worker release against NSS.
# No parameters are needed, unlike the gecko try script.
######

function open_browser_page {
  case "$OSTYPE" in
    linux*)
      xdg-open "${1}"
      ;;
    darwin*)
      open "${1}"
      ;;
  esac
}

cd "$(dirname "${0}")"
THIS_SCRIPT_DIR="$(pwd)"

NEW_VERSION="$(cat ../worker-type-host-definitions/aws-provisioner-v1/nss-win2012r2-new/userdata | sed -n 's_.*https://github\.com/taskcluster/generic-worker/releases/download/v\(.*\)/generic-worker-windows-amd64\.exe.*_\1_p')"
VALID_FORMAT='^[1-9][0-9]*\.\(0\|[1-9][0-9]*\)\.\(0\|[1-9]\)\([0-9]*alpha[1-9][0-9]*\|[0-9]*\)$'

if ! echo "${NEW_VERSION}" | grep -q "${VALID_FORMAT}"; then
  echo "Release version '${NEW_VERSION}' not allowed" >&2
  exit 65
fi

export WORKER_TYPES_DIR=../worker-type-host-definitions/aws-provisioner-v1
../../aws/update-worker-types/worker_type.sh nss-win2012r2-new update

NSS_CHECKOUT="$(mktemp -d -t nss-checkout.XXXXXXXXXX)"
cd "${NSS_CHECKOUT}"
hg clone https://hg.mozilla.org/projects/nss
cd nss
grep -rl nss-win2012r2 . | while read FILE
do
  cp "${FILE}" "${FILE}.x"
  cat "${FILE}.x" | sed 's/nss-win2012r2/nss-win2012r2-new/g' > "${FILE}"
  rm "${FILE}.x"
done
hg commit -m "Testing generic-worker ${NEW_VERSION} on nss-win2012r2-new worker type; try: -p win,win64 -t none -u all"
hg push -f ssh://hg.mozilla.org/projects/nss-try -r .

open_browser_page 'https://treeherder.mozilla.org/#/jobs?repo=nss-try'

cd "${THIS_SCRIPT_DIR}"
rm -rf "${NSS_CHECKOUT}"
