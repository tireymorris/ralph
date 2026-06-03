#!/usr/bin/env bash
set -euo pipefail

REPO="${RALPH_REPO:-https://github.com/tireymorris/ralph.git}"
REF="${1:-main}"

tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t ralph-install)
trap 'rm -rf "$tmpdir"' EXIT

git clone --depth 1 --branch "$REF" "$REPO" "$tmpdir/ralph"
(cd "$tmpdir/ralph" && go install .)

bindir="$(go env GOPATH)/bin"
echo "installed ${bindir}/ralph"
