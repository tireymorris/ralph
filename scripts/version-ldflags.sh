#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

version="$(git describe --tags --always 2>/dev/null || echo dev)"
commit="$(git rev-parse HEAD)"
ref="$(git symbolic-ref -q --short HEAD 2>/dev/null || git describe --tags --exact-match HEAD 2>/dev/null || echo unknown)"

printf '%s' "-X github.com/tireymorris/ralph/internal/version.Version=${version} -X github.com/tireymorris/ralph/internal/version.Commit=${commit} -X github.com/tireymorris/ralph/internal/version.Ref=${ref}"
