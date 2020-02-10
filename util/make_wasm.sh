#!/bin/sh
set -eu

extradir="$(dirname "$0")/../extra"

mkdir -pv wasm

env GOOS=js GOARCH=wasm go build -o wasm/main.wasm -v -tags noclipboard "$@"

cp -v "$(go env GOROOT)/misc/wasm/wasm_exec.js" wasm/
cp -v "$extradir/index.html" wasm/

# vim: set sw=2 et :
