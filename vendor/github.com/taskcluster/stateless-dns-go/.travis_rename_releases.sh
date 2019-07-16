#!/bin/bash

# The deploy is called per arch and os combination - so we only release one file here.
# We just need to work out which file we built, rename it to something unique, and
# set an environment variable for its location that we can use in .travis.yml for
# publishing back to github.

# all cross-compiled binaries are in subdirectories: ${GOPATH}/bin/${GOOS}_${GOARCH}/
# linux 64 bit, not cross-compiled, breaks this rule and is in ${GOPATH}/bin
# therefore move it to match the convention of the others, to simplify subsequent steps
# note: we don't know what we built, so only move it if we happen to be linux amd64 travis job
if [ -f "${GOPATH}/bin/create-hostname" ]; then
  mkdir "${GOPATH}/bin/linux_amd64"
  mv "${GOPATH}/bin/create-hostname" "${GOPATH}/bin/linux_amd64/create-hostname"
  mv "${GOPATH}/bin/decode-hostname" "${GOPATH}/bin/linux_amd64/decode-hostname"
fi

# linux, darwin:
FILE_EXT=""
[ "${GOOS}" == "windows" ] && FILE_EXT=".exe"

# let's rename the release file because it has a 1:1 mapping with what it is called on
# github releases, and therefore the name for each platform needs to be unique so that
# they don't overwrite each other. Set a variable that can be used in .travis.yml
export CREATE_HOSTNAME_RELEASE_FILE="${TRAVIS_BUILD_DIR}/create-hostname-${GOOS}-${GOARCH}${FILE_EXT}"
export DECODE_HOSTNAME_RELEASE_FILE="${TRAVIS_BUILD_DIR}/decode-hostname-${GOOS}-${GOARCH}${FILE_EXT}"
mv "${GOPATH}/bin/${GOOS}_${GOARCH}/create-hostname${FILE_EXT}" "${CREATE_HOSTNAME_RELEASE_FILE}"
mv "${GOPATH}/bin/${GOOS}_${GOARCH}/decode-hostname${FILE_EXT}" "${DECODE_HOSTNAME_RELEASE_FILE}"
