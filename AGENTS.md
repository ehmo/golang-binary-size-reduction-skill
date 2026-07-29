# Go Binary Size Reduction

Shrink Go binaries with measured tradeoffs. Covers CLIs, daemons, libraries, plugins, Wasm, and packaged apps.

**References:** https://pkg.go.dev/cmd/go and https://pkg.go.dev/golang.org/x/tools/go/analysis

## Technique priority

| # | Technique | Typical raw win | Risk |
|---|-----------|----------------|------|
| 1 | `-ldflags="-s -w"` (strip symbols/DWARF) | 25-35% | Weaker debugging |
| 2 | `-gcflags=all=-l` (disable inlining) | 5-10% additional | Reduced runtime performance |
| 3 | `CGO_ENABLED=0` + `-tags netgo,osusergo` | 0-6% | Behavior changes; can increase size |
| 3b | `CGO_CFLAGS="-Oz"` (when CGO must stay on) | 1-2% | Slightly slower C code; no effect if no C source compiled |
| 4 | Project-specific build tags | 0-15% | Feature removal |
| 5 | Dependency pruning / embedded asset reduction | Varies | Source changes required |
| 6 | `garble -tiny`, UPX, TinyGo | Varies | Specialist tradeoffs |

## Hard rules

- Never apply `-gcflags=all=-B` (disables bounds checks)
- Never run UPX on macOS (kernel kills packed binaries)
- Never pack or patch after code signing
- Never treat PGO as a size reduction technique
- Always measure raw size, gzip size, and runtime behavior
- Always collect build context before changing anything
- Treat analyzer findings as audit leads, not automatic edits

## Scripts

- `scripts/collect-build-context.sh` -- Gather build facts
- `scripts/analyze-project.sh` -- Run source analysis across a Go project
- `scripts/reproducible-build.sh` -- Build with stable flags
- `scripts/measure-binary-size.sh` -- Measure artifact sizes
- `scripts/compare-size-report.sh` -- Diff two artifacts

`collect-build-context.sh` runs the analyzer over non-test files in `./...`. Analyzer exit status 3 means it completed and reported findings.
