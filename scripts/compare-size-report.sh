#!/usr/bin/env bash
set -euo pipefail

if (($# != 2)); then
  printf 'usage: compare-size-report.sh <before-artifact> <after-artifact>\n' >&2
  exit 64
fi

before="$1"
after="$2"

for artifact in "$before" "$after"; do
  if [[ ! -f "$artifact" ]]; then
    printf 'artifact not found: %s\n' "$artifact" >&2
    exit 66
  fi
done

file_bytes() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    stat -f '%z' "$1"
  else
    stat -c '%s' "$1"
  fi
}

gzip_bytes() {
  gzip -n -9 -c "$1" | wc -c | tr -d ' '
}

xz_bytes() {
  if command -v xz >/dev/null 2>&1; then
    xz -9e -c "$1" | wc -c | tr -d ' '
  else
    printf '0\n'
  fi
}

delta_pct() {
  awk -v before="$1" -v after="$2" 'BEGIN {
    if (before == 0) {
      print "0.00"
      exit
    }
    printf "%.2f", ((after - before) / before) * 100
  }'
}

section() {
  printf '== %s ==\n' "$1"
}

before_bytes="$(file_bytes "$before")"
after_bytes="$(file_bytes "$after")"
before_gzip="$(gzip_bytes "$before")"
after_gzip="$(gzip_bytes "$after")"
before_xz="$(xz_bytes "$before")"
after_xz="$(xz_bytes "$after")"

section "summary"
printf 'before=%s\n' "$before"
printf 'after=%s\n' "$after"
printf 'before_bytes=%s\n' "$before_bytes"
printf 'after_bytes=%s\n' "$after_bytes"
printf 'delta_bytes=%s\n' "$((after_bytes - before_bytes))"
printf 'delta_pct=%s\n' "$(delta_pct "$before_bytes" "$after_bytes")"
printf 'before_gzip_bytes=%s\n' "$before_gzip"
printf 'after_gzip_bytes=%s\n' "$after_gzip"
printf 'delta_gzip_bytes=%s\n' "$((after_gzip - before_gzip))"
printf 'delta_gzip_pct=%s\n' "$(delta_pct "$before_gzip" "$after_gzip")"
printf 'before_xz_bytes=%s\n' "$before_xz"
printf 'after_xz_bytes=%s\n' "$after_xz"
printf 'delta_xz_bytes=%s\n' "$((after_xz - before_xz))"
printf 'delta_xz_pct=%s\n' "$(delta_pct "$before_xz" "$after_xz")"

section "go-version-before"
go version -m "$before" 2>&1 || true

section "go-version-after"
go version -m "$after" 2>&1 || true
