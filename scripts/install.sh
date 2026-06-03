#!/usr/bin/env bash
set -euo pipefail

REPO="${RALPH_REPO:-https://github.com/tireymorris/ralph.git}"
REF="${1:-${RALPH_REF:-main}}"
MIN_GO_MAJOR=1
MIN_GO_MINOR=24
RALPH_INSTALL_DEPS="${RALPH_INSTALL_DEPS:-1}"
RALPH_GO_ROOT="${RALPH_GO_ROOT:-${HOME}/.local/go}"

die() {
  echo "error: $*" >&2
  exit 1
}

go_version_ok() {
  command -v go >/dev/null 2>&1 || return 1
  local ver go_major rest go_minor
  ver=$(go env GOVERSION | sed 's/^go//')
  go_major=${ver%%.*}
  rest=${ver#*.}
  go_minor=${rest%%.*}
  [ "$go_major" -gt "$MIN_GO_MAJOR" ] 2>/dev/null && return 0
  [ "$go_major" -eq "$MIN_GO_MAJOR" ] && [ "$go_minor" -ge "$MIN_GO_MINOR" ] 2>/dev/null
}

detect_goos() {
  case "$(uname -s 2>/dev/null || true)" in
    Darwin) printf '%s' "darwin" ;;
    Linux) printf '%s' "linux" ;;
    MINGW*|MSYS*|CYGWIN*) die "native Windows shell is unsupported; use WSL or Git Bash and retry" ;;
    *) die "unsupported OS; install Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+ from https://go.dev/dl/" ;;
  esac
}

detect_goarch() {
  case "$(uname -m 2>/dev/null || true)" in
    x86_64|amd64) printf '%s' "amd64" ;;
    aarch64|arm64) printf '%s' "arm64" ;;
    *) die "unsupported CPU; install Go from https://go.dev/dl/" ;;
  esac
}

install_go_user_local() {
  local goos goarch archive url version tmpdir
  goos=$(detect_goos)
  goarch=$(detect_goarch)
  command -v curl >/dev/null 2>&1 || command -v wget >/dev/null 2>&1 || \
    die "curl or wget is required to download Go"
  version=$(curl -fsSL https://go.dev/VERSION?m=text | sed 's/^go//')
  archive="go${version}.${goos}-${goarch}.tar.gz"
  url="https://go.dev/dl/${archive}"
  tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t ralph-go)
  echo "installing Go ${version} to ${RALPH_GO_ROOT}..."
  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "${tmpdir}/${archive}"
  else
    wget -qO "${tmpdir}/${archive}" "$url"
  fi
  rm -rf "${RALPH_GO_ROOT}"
  mkdir -p "$(dirname "${RALPH_GO_ROOT}")"
  tar -C "$(dirname "${RALPH_GO_ROOT}")" -xzf "${tmpdir}/${archive}"
  rm -rf "$tmpdir"
  export PATH="${RALPH_GO_ROOT}/bin:${PATH}"
}

try_brew_install() {
  local pkg=$1
  command -v brew >/dev/null 2>&1 || return 1
  echo "installing ${pkg} via Homebrew..."
  brew install "$pkg"
}

ensure_git() {
  command -v git >/dev/null 2>&1 && return 0
  [ "$RALPH_INSTALL_DEPS" = "1" ] || die "git is required (https://git-scm.com/downloads)"
  try_brew_install git || die "git is required (https://git-scm.com/downloads)"
  command -v git >/dev/null 2>&1 || die "git install failed"
}

ensure_go() {
  go_version_ok && return 0
  if [ "$RALPH_INSTALL_DEPS" != "1" ]; then
    if command -v go >/dev/null 2>&1; then
      die "Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+ required (found $(go env GOVERSION)); set RALPH_INSTALL_DEPS=1 to auto-install"
    fi
    die "Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+ required (https://go.dev/dl/); set RALPH_INSTALL_DEPS=1 to auto-install"
  fi
  if command -v go >/dev/null 2>&1; then
    die "Go ${MIN_GO_MAJOR}.${MIN_GO_MINOR}+ required (found $(go env GOVERSION)); upgrade Go manually"
  fi
  try_brew_install go || true
  go_version_ok && return 0
  install_go_user_local
  go_version_ok || die "Go install failed; add to PATH: export PATH=\"${RALPH_GO_ROOT}/bin:\$PATH\""
}

resolve_bindir() {
  if [ -n "${RALPH_BIN_DIR:-}" ]; then
    printf '%s' "$RALPH_BIN_DIR"
    return
  fi

  local gobin gopath
  gobin=$(go env GOBIN 2>/dev/null || true)
  if [ -n "$gobin" ]; then
    printf '%s' "$gobin"
    return
  fi

  gopath=$(go env GOPATH)
  case "$gopath" in
    *:*) gopath=${gopath%%:*} ;;
  esac
  case "$gopath" in
    */packages) printf '%s' "${gopath%/packages}/bin" ;;
    *) printf '%s' "${gopath}/bin" ;;
  esac
}

ensure_git
ensure_go

tmpdir=$(mktemp -d 2>/dev/null || mktemp -d -t ralph-install)
cleanup() { rm -rf "$tmpdir"; }
trap cleanup EXIT

bindir=$(resolve_bindir)
mkdir -p "$bindir"
target="${bindir}/ralph"

echo "installing ralph from ${REPO}@${REF}..."
git clone --depth 1 --branch "$REF" "$REPO" "$tmpdir/ralph"
(
  cd "$tmpdir/ralph"
  GOBIN="$bindir" go install .
)

[[ -x "$target" ]] || die "install failed: ${target} not found (GOPATH=$(go env GOPATH))"

if command -v ralph >/dev/null 2>&1; then
  echo "installed $(command -v ralph)"
  ralph --help | head -n 1
else
  echo "installed ${target}"
  echo "add to PATH: export PATH=\"${bindir}:\$PATH\""
fi

case ":${PATH}:" in
  *":${RALPH_GO_ROOT}/bin:"*) ;;
  *)
    if [ -x "${RALPH_GO_ROOT}/bin/go" ]; then
      echo "add Go to PATH: export PATH=\"${RALPH_GO_ROOT}/bin:\$PATH\""
    fi
    ;;
esac

echo "requires a runner on PATH: claude (default), cursor-agent, opencode, or pi"
