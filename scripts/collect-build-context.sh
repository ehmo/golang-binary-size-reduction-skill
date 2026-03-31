#!/usr/bin/env bash
set -euo pipefail

pkg="${1:-.}"

section() {
  printf '== %s ==\n' "$1"
}

watchlist_pattern='^(plugin|reflect|text/template|html/template|embed|time/tzdata|os/user|net)$'
tmp="$(mktemp)"
trap 'rm -f "$tmp"' EXIT

section "target"
printf 'package=%s\n' "$pkg"

section "go-version"
go version

section "go-env-json"
go env -json GOOS GOARCH GOAMD64 GOARM GO386 GOEXPERIMENT CGO_ENABLED CC CXX GOMOD GOWORK GOFLAGS GOROOT GOPATH GOMODCACHE

section "package-json"
go list -json "$pkg"

section "deps-non-std"
go list -deps -f '{{if not .Standard}}{{.ImportPath}}{{end}}' "$pkg" | LC_ALL=C sort -u

section "watchlist-hits"
go list -deps -f '{{.ImportPath}}' "$pkg" | LC_ALL=C sort -u > "$tmp"
if command -v rg >/dev/null 2>&1; then
  rg -n "$watchlist_pattern" "$tmp" || true
else
  grep -nE "$watchlist_pattern" "$tmp" || true
fi
