#!/usr/bin/env bash
set -euo pipefail

if (($# != 1)); then
  printf 'usage: measure-binary-size.sh <artifact>\n' >&2
  exit 64
fi

artifact="$1"

if [[ ! -f "$artifact" ]]; then
  printf 'artifact not found: %s\n' "$artifact" >&2
  exit 66
fi

section() {
  printf '== %s ==\n' "$1"
}

file_bytes() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    stat -f '%z' "$1"
  else
    stat -c '%s' "$1"
  fi
}

sha256_file() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | awk '{print $1}'
  else
    sha256sum "$1" | awk '{print $1}'
  fi
}

gzip_bytes() {
  gzip -n -9 -c "$1" | wc -c | tr -d ' '
}

xz_bytes() {
  if command -v xz >/dev/null 2>&1; then
    xz -9e -c "$1" | wc -c | tr -d ' '
  else
    printf 'unavailable\n'
  fi
}

section "summary"
printf 'path=%s\n' "$artifact"
printf 'bytes=%s\n' "$(file_bytes "$artifact")"
printf 'sha256=%s\n' "$(sha256_file "$artifact")"
printf 'gzip_bytes=%s\n' "$(gzip_bytes "$artifact")"
printf 'xz_bytes=%s\n' "$(xz_bytes "$artifact")"

section "file"
file "$artifact" 2>/dev/null || true

section "go-version-m"
go version -m "$artifact" 2>&1 || true

section "top-symbols"
go tool nm -size -sort size "$artifact" 2>/dev/null | head -n 40 || true
