#!/usr/bin/env bash
set -euo pipefail

# Only supporting tabular here
readonly TABULAR_PKG=go.pennock.tech/tabular

cd "$(dirname "$0")/.."

declare -a go_ldflags go_tags
go_ldflags=()
go_tags=()

go_ldflags+=( -X "$(go list ./commands/version).VersionString=$(./util/repo_version)" )

TABULAR_VERSION_VAR="$(go list -m -f '{{.Path}}' "$TABULAR_PKG").LinkerSpecifiedVersion"
TABULAR_VERSION_VALUE="$(go list -m -f '{{.Version}}' "$TABULAR_PKG")"
go_ldflags+=( -X "$TABULAR_VERSION_VAR=$TABULAR_VERSION_VALUE" )

cmdline=(
    "${GO_CMD:-go}" build -o character
    )

if [[ "${#go_ldflags[@]}" -gt 0 ]]; then
  cmdline+=(-ldflags "${go_ldflags[*]}")
fi
if [[ "${#go_tags[@]}" -gt 0 ]]; then
  cmdline+=(-tags "${go_tags[*]}")
fi

set -x
"${cmdline[@]}"

# vim: set sw=2 et :
