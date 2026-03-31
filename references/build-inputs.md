# Build Inputs

Collect these facts before proposing shrink work.

## Artifact Facts

- target package or packages
- current output artifact path
- target OS and architecture
- buildmode
- whether the artifact is a CLI, daemon, plugin, shared library, Wasm module, or packaged app

## Toolchain Facts

- Go version
- exact build command
- `GOFLAGS`
- `CGO_ENABLED`
- explicit `-tags`, `-ldflags`, and `-gcflags`
- whether the repo depends on specific Go version behavior

## Runtime Facts

- whether cgo is required
- whether `plugin` is required
- whether `net` must use native resolver behavior
- whether `os/user` must use libc-backed lookups
- whether embedded timezone data is required
- whether large assets are embedded with `embed`

## Release Facts

- whether symbols and DWARF are needed in the shipped artifact
- whether crash symbolization happens from the shipped binary or from a retained companion artifact
- whether the target is signed or notarized
- whether the artifact is further compressed by the delivery path
- whether startup latency or cold-start time matters

## Dependency Facts

- non-standard-library dependency list
- direct imports in `main`
- presence of heavy SDKs or optional integrations
- use of `reflect`, `text/template`, `html/template`, `plugin`, `embed`, `time/tzdata`, `net`, `os/user`

## Agent Notes

- Prefer collecting facts with `./scripts/collect-build-context.sh`.
- If the build already uses tags or unusual env vars, preserve them during measurement.
- If the release process signs binaries, perform any packing or patching analysis before signing, not after.
