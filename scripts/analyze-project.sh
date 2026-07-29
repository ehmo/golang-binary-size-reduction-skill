#!/usr/bin/env bash
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
skill_root="$(cd -- "$script_dir/.." && pwd)"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/gosizeaudit.XXXXXX")"
trap 'rm -rf -- "$tmp_dir"' EXIT
host_goos="$(go env GOHOSTOS)"
host_goarch="$(go env GOHOSTARCH)"

if (($# == 0)); then
  set -- ./...
fi

(
  cd "$skill_root"
  env \
    GOOS="$host_goos" \
    GOARCH="$host_goarch" \
    CGO_ENABLED=0 \
    GOEXPERIMENT= \
    GOFLAGS= \
    GOWORK=off \
    go build -trimpath -o "$tmp_dir/gosizeaudit" ./cmd/gosizeaudit
)

"$tmp_dir/gosizeaudit" -test=false "$@"
