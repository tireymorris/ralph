#!/usr/bin/env bash
set -euo pipefail

REPO="${RALPH_REPO:-https://github.com/tireymorris/ralph.git}"
REF="${1:-${RALPH_REF:-main}}"
MIN_GO_MAJOR=1
MIN_GO_MINOR=24

die() {
  echo "error: $*" >&2
  exit 1
}

command -v git >/dev/null 2>&1 || die "git is required"

command -v go >/dev/null 2>&1 || die "Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+ is required (https://go.dev/dl/)"

ver=$(go env GOVERSION | sed 's/^go//')
go_major=${ver%%.*}
rest=${ver#*.}
go_minor=${rest%%.*}
if [ "$go_major" -lt "$MIN_GO_MAJOR" ] 2>/dev/null || { [ "$go_major" -eq "$MIN_GO_MAJOR" ] && [ "$go_minor" -lt "$MIN_GO_MINOR" ]; }; then
  die "Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+ is required (found $(go env GOVERSION))"
fi

tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t ralph-install)
cleanup() { rm -rf "$tmpdir"; }
trap cleanup EXIT

echo "installing ralph from ${REPO}@${REF}..."
git clone --depth 1 --branch "$REF" "$REPO" "$tmpdir/ralph"
(
  cd "$tmpdir/ralph"
  go install .
)

bindir="$(go env GOPATH)/bin"
target="${bindir}/ralph"

[[ -x "$target" ]] || die "install failed: ${target} not found"

if command -v ralph >/dev/null 2>&1; then
  echo "installed $(command -v ralph)"
  ralph --help | head -n 1
else
  echo "installed ${target}"
  echo "add to PATH: export PATH=\"${bindir}:\$PATH\""
fi

echo "requires a runner on PATH: claude (default), cursor-agent, opencode, or pi"
