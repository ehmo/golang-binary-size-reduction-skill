# Changelog

## 1.1.0 — 2026-04-01

### Added

- `CGO_CFLAGS="-Oz"` as opt-in technique for CGO-enabled projects that compile C source code (e.g., mattn/go-sqlite3). Benchmarked 1-2% raw size savings (~500 KB) on owncast, waveterm, and authelia. No effect when CGO only links system libraries.
- `--cgo-cflags` flag in `scripts/reproducible-build.sh`.
- Verification checklist for post-CGO_CFLAGS changes.
- Benchmark data for `CGO_CFLAGS="-Oz"` and `CC="zig cc"` in `references/sources.md`.
- `CC="zig cc"` documented as a non-technique in the decision tree (0-0.8% size *increase* on native builds; it is a cross-compilation tool, not a size tool).

### Changed

- Decision tree: added CGO_CFLAGS opt-in row, zig cc do-not-use row, and CGO branch guidance for C source optimization.
- Workflow Phase 4: added CGO_CFLAGS block with example command when CGO must stay enabled.
- AGENTS.md technique table: added row 3b for CGO_CFLAGS.
- SKILL.md: expanded step 3 (runtime reductions) and fast triage with CGO_CFLAGS and zig cc guidance.

## 1.0.0 — 2026-03-30

Initial release. Tested against 14 trending Go repositories with 36.6% average raw size reduction.
