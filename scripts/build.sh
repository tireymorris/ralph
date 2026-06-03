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

ldflags="$("${root}/scripts/version-ldflags.sh")"
go build -ldflags "$ldflags" -o "$out" .
