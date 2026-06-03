#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

out=""
while [[ $# -gt 0 ]]; do
  case "$1" in
    -o)
      out="$2"
      shift 2
      ;;
    *)
      echo "usage: $(basename "$0") -o <path>" >&2
      exit 1
      ;;
  esac
done

if [[ -z "$out" ]]; then
  echo "usage: $(basename "$0") -o <path>" >&2
  exit 1
fi

version="$(git describe --tags --always 2>/dev/null || echo dev)"
commit="$(git rev-parse HEAD)"
ref="$(git symbolic-ref -q --short HEAD 2>/dev/null || git describe --tags --exact-match HEAD 2>/dev/null || echo unknown)"

ldflags=(
  "-X" "ralph/internal/version.Version=${version}"
  "-X" "ralph/internal/version.Commit=${commit}"
  "-X" "ralph/internal/version.Ref=${ref}"
)

go build -ldflags "${ldflags[*]}" -o "$out" .
