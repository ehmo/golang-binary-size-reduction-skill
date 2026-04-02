# Verification

A smaller binary is not a valid result until it passes both artifact checks and behavior checks.

## Artifact Checks

For every before/after pair, capture:

1. raw bytes
2. gzip bytes
3. xz bytes when available
4. `go version -m` output
5. top symbols from `go tool nm -size -sort size`

Use:

```bash
./scripts/measure-binary-size.sh dist/app
./scripts/compare-size-report.sh dist/app-before dist/app-after
```

## Behavior Checks

Run the smallest useful set of runtime checks for the target:

1. process starts successfully
2. main request path or CLI command still works
3. panic or crash output remains acceptable for the release policy
4. startup latency and RSS remain acceptable if a packer or alternate runtime was used

## Special Checks

### After `-ldflags="-s -w"`

- confirm the release process still has adequate crash-symbolization support
- keep an unstripped companion artifact if the org needs postmortem debugging

### After `CGO_CFLAGS="-Oz"`

- verify C-code-intensive paths still meet performance requirements
- this only affects C source compilation, not system library linking
- safe for most projects since C code (e.g., SQLite) is rarely in the hot path

### After `CGO_ENABLED=0`, `netgo`, or `osusergo`

- test DNS resolution behavior
- test user and group lookup behavior if used
- confirm certificates, proxy behavior, and environment-driven resolver behavior still match expectations

### After removing `timetzdata`

- test timezone-sensitive paths on systems that may not have system tzdata installed

### After reducing embedded assets

- test asset loading, templates, and static-file serving

### After `-gcflags=all=-l`

- verify performance-critical paths still meet latency requirements
- benchmark hot loops if applicable
- acceptable for CLIs, build tools, and size-critical deployments
- less suitable for latency-sensitive services or inner-loop compute

### After UPX

- only on Linux/Windows — macOS kills packed binaries (SIGKILL)
- test cold start
- test RSS during startup
- test antivirus, malware scanning, and unpacking workflows if relevant
- verify signing only after final packed artifact exists

### After TinyGo

- test feature compatibility, runtime assumptions, and target-specific behavior

### After macOS changes

- re-run codesigning checks
- re-run notarization checks if the distribution requires them

## Release Gate

Ship only if:

1. the measured win is real
2. the runtime behavior is still acceptable
3. debugging and provenance tradeoffs are documented
4. signing and packaging constraints still pass
