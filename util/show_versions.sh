#!/bin/sh
# Not -eu: we want to keep going and show as much as possible, even if something fails
set +eu

cd "$(dirname "$0")"
cd ..
set -x

date
uname -a
git version
go version
./util/repo_version
go mod download -json | jq -r '"\(.Path)\t\(.Version)"'
go mod verify
