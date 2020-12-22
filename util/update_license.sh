#!/usr/bin/env bash
set -euo pipefail

progname="$(basename "$0" .sh)"
warn() { printf >&2 '%s: %s\n' "$progname" "$*"; }
die() { warn "$@"; exit 1; }
hereis="$(pwd)"

readonly OUTFILE=LICENSES_all.txt
readonly tmpstage="$hereis/tmplicpart"
readonly tmpaccumulate="$hereis/tmplic"

for require in go.mod LICENSE.txt; do
  test -f "$require" || die "missing $require in current directory"
done

rm -f -- "$OUTFILE" "$tmpstage" "$tmpaccumulate"

go mod download -json | jq -r '"\(.Path):\(.Dir)"' | while read pair; do
  modpath="${pair%%:*}"
  moddir="${pair#*:}"
  (
    cd "$moddir" || die "can't chdir($moddir) for [$modpath]"
    for F in NOTICE* LICEN[SC]E* PATENTS; do
      test -s "$F" || continue
      echo "~~~ $F - $modpath ~~~"
      cat "./$F"
      # shell strips final newline and we're looking for a newline, so
      # really this will end up as the empty string when the file ends in a
      # POSIX full line:
      last_char="$(tail -c 1 "./$F")"
      if [[ "$last_char" != $'\n' && "$last_char" != "" ]]; then
        echo
      fi
    done
  ) > "$tmpstage"
  test -s "$tmpstage" || continue
  echo "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~"
  cat -- "$tmpstage"
  echo
done > "$tmpaccumulate"

(
  echo "~~~ $(go list .) ~~~"
  cat LICENSE.txt "$tmpaccumulate"
) > "$OUTFILE"

rm -f -- "$tmpstage" "$tmpaccumulate"

# vim: set sw=2 et :
