#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
usage: reproducible-build.sh -o <artifact> [options]

Options:
  --pkg <package>          package to build (default: .)
  -o, --output <path>      output artifact path
  --strip                  add -ldflags="-s -w"
  --omit-vcs-stamp         add -buildvcs=false
  --disable-cgo            set CGO_ENABLED=0 for the build
  --tags <list>            pass -tags
  --buildmode <mode>       pass -buildmode
  --ldflags <flags>        append linker flags
  --gcflags <flags>        pass gcflags
  --cgo-cflags <flags>     set CGO_CFLAGS (e.g., -Oz for size optimization)
  -- <extra args...>       extra go build args before the package
EOF
}

pkg="."
out=""
strip=0
omit_vcs=0
disable_cgo=0
tags=""
buildmode=""
ldflags=""
gcflags=""
cgo_cflags=""
extra_args=()

while (($# > 0)); do
  case "$1" in
    --pkg)
      pkg="$2"
      shift 2
      ;;
    -o|--output)
      out="$2"
      shift 2
      ;;
    --strip)
      strip=1
      shift
      ;;
    --omit-vcs-stamp)
      omit_vcs=1
      shift
      ;;
    --disable-cgo)
      disable_cgo=1
      shift
      ;;
    --tags)
      tags="$2"
      shift 2
      ;;
    --buildmode)
      buildmode="$2"
      shift 2
      ;;
    --ldflags)
      ldflags="$2"
      shift 2
      ;;
    --gcflags)
      gcflags="$2"
      shift 2
      ;;
    --cgo-cflags)
      cgo_cflags="$2"
      shift 2
      ;;
    --)
      shift
      extra_args=("$@")
      break
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      printf 'unknown argument: %s\n' "$1" >&2
      usage >&2
      exit 64
      ;;
  esac
done

if [[ -z "$out" ]]; then
  printf 'missing required output path\n' >&2
  usage >&2
  exit 64
fi

mkdir -p "$(dirname "$out")"

go_args=(build -trimpath -o "$out")

if [[ "$omit_vcs" -eq 1 ]]; then
  go_args+=(-buildvcs=false)
fi

if [[ -n "$tags" ]]; then
  go_args+=(-tags "$tags")
fi

if [[ -n "$buildmode" ]]; then
  go_args+=(-buildmode "$buildmode")
fi

final_ldflags="$ldflags"
if [[ "$strip" -eq 1 ]]; then
  if [[ -n "$final_ldflags" ]]; then
    final_ldflags+=" "
  fi
  final_ldflags+="-s -w"
fi

if [[ -n "$final_ldflags" ]]; then
  go_args+=(-ldflags "$final_ldflags")
fi

if [[ -n "$gcflags" ]]; then
  go_args+=(-gcflags "$gcflags")
fi

if ((${#extra_args[@]} > 0)); then
  go_args+=("${extra_args[@]}")
fi

go_args+=("$pkg")

effective_cgo="inherit"
if [[ "$disable_cgo" -eq 1 ]]; then
  effective_cgo="0"
fi

{
  printf 'CGO_ENABLED=%s\n' "$effective_cgo"
  if [[ -n "$cgo_cflags" ]]; then
    printf 'CGO_CFLAGS=%s\n' "$cgo_cflags"
  fi
  printf 'command='
  printf '%q ' go "${go_args[@]}"
  printf '\n'
} >&2

build_env=()
if [[ "$disable_cgo" -eq 1 ]]; then
  build_env+=(CGO_ENABLED=0)
fi
if [[ -n "$cgo_cflags" ]]; then
  build_env+=(CGO_CFLAGS="$cgo_cflags")
fi

if ((${#build_env[@]} > 0)); then
  env "${build_env[@]}" go "${go_args[@]}"
else
  go "${go_args[@]}"
fi
