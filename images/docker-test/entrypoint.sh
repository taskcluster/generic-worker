#!/bin/bash -e

export TASKCLUSTER_CLIENT_ID=$(curl -s http://taskcluster/secrets/v1/secret/repo:github.com/taskcluster/docker-worker:ci-creds | jq -r '.secret.client_id')
export TASKCLUSTER_ACCESS_TOKEN=$(curl -s http://taskcluster/secrets/v1/secret/repo:github.com/taskcluster/docker-worker:ci-creds | jq -r '.secret.access_token')

repo_url=$1
repo_sha=$2
go_package="$(url-parser --part=hostname $repo_url)/$(url-parser --part=path $repo_url)"
go_package="${go_package%%.git}"
tags=docker

go get $go_package

cd $GOPATH/src/$go_package

git checkout $repo_sha

go get -t -tags=$tags ./...
go test -v -tags=$tags ./...
