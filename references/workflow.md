# Workflow

Use this workflow unless the repo already has a stricter release pipeline.

## Phase 1: Baseline

1. Identify the target package or artifact.
2. Capture the build context:
   `./scripts/collect-build-context.sh ./cmd/app`
3. Produce a baseline artifact with stable paths:
   `./scripts/reproducible-build.sh -o dist/app-baseline --pkg ./cmd/app`
4. Measure it:
   `./scripts/measure-binary-size.sh dist/app-baseline`
5. Review `source-size-analysis` in the collected context. The diagnostics identify audit leads in active and ignored Go files. They do not prove that code reaches the target binary or that removing it is safe.

Do not change code or flags before you have a baseline.

## Phase 2: Cheap Safe Wins

Try these first. Each step typically builds on the previous:

1. `-trimpath` (good hygiene, small win)
2. `-buildvcs=false` (omit VCS stamp)
3. `-ldflags="-s -w"` (strip symbols and DWARF — this is the biggest single win, typically 25-35% raw reduction, 45-60% compressed)
4. `-gcflags=all=-l` (disable inlining — adds 5-10pp of raw reduction on top of stripping; tradeoff is reduced runtime performance)

Example:

```bash
./scripts/reproducible-build.sh \
  -o dist/app-stripped \
  --pkg ./cmd/app \
  --omit-vcs-stamp \
  --strip \
  --gcflags "all=-l"
./scripts/compare-size-report.sh dist/app-baseline dist/app-stripped
```

Measured benchmark results (14 trending Go repos):
- Strip alone: avg 29% raw reduction
- Strip + gcflags=-l: avg 36% raw reduction
- Strip + gcflags=-l + CGO=0 + tags: avg 36-46% raw reduction (varies by repo)

If the win is large enough and the tradeoffs are acceptable, stop there.

## Phase 2b: Repo-Specific Tag Discovery

Before structural changes, check the project's source and build infrastructure for existing feature-gating tags:

1. Review the analyzer's `build-tag` findings for active and ignored Go files.
2. Read `goreleaser.yaml`, `Makefile`, `Taskfile.yaml`, `Dockerfile`, and CI workflow files.
3. Search for `-tags` flags in build commands.
4. Search source for `//go:build` constraints if the analyzer could not load every shipped package.
5. Common patterns: `WITHOUT_DOCKER`, `production`, `nodynamic`, `sqlite_omit_load_extension`.
6. These tags can remove entire dependency subtrees, often producing 5-15% additional wins.

## Phase 3: Structural Reductions

If stripping and tags do not move size enough, inspect structure instead of reaching for exotic flags.

Priorities:

1. Review `binsize` diagnostics for side-effect imports and packages that may retain runtime machinery.
2. Remove unused direct imports from `main`.
3. Move optional features behind build tags.
4. Split heavyweight integrations into separate subcommands or binaries.
5. Reduce or externalize embedded assets.
6. Remove `timetzdata` unless the target truly needs embedded timezone data.

This phase usually delivers the biggest durable wins for projects that accept source changes.

## Phase 4: Runtime and Platform Reductions

Default to `CGO_ENABLED=0` with `-tags netgo,osusergo`. These three settings always go together:

```bash
CGO_ENABLED=0 go build -trimpath -tags netgo,osusergo -ldflags="-s -w" -gcflags="all=-l" -o dist/app ./cmd/app
```

If the resulting binary is *larger* than the CGO-enabled build, revert to CGO enabled. This is rare but happens when C library implementations are smaller than their pure-Go replacements.

Check for cgo requirements before applying:
1. Run `go env CGO_ENABLED` to see the default
2. Check the project's goreleaser.yaml or Dockerfile for explicit `CGO_ENABLED=1`
3. If the project uses `mattn/go-sqlite3`, `cgo`-backed crypto, or external C libraries, CGO=0 may break the build or change behavior

Do not apply CGO=0 without netgo and osusergo — without these tags, the runtime may still try to use cgo-backed resolver or user lookup functions and fail.

If CGO must stay enabled and the project compiles C source code (e.g., `mattn/go-sqlite3` amalgamation), apply `CGO_CFLAGS="-Oz"`:

```bash
CGO_ENABLED=1 CGO_CFLAGS="-Oz" go build -trimpath -ldflags="-s -w" -gcflags="all=-l" -o dist/app ./cmd/app
```

This optimizes the C compiler for size instead of speed. Benchmarked savings: ~500 KB / 1-2% raw on SQLite-based projects (owncast, waveterm, authelia). Has no effect when CGO only links system libraries without compiling C source (e.g., wails linking Cocoa/WebKit).

## Phase 5: Specialist Tracks

Use only when the earlier phases are exhausted or the distribution model clearly supports them.

1. `garble -tiny`
2. UPX (Linux/Windows only — **does not work on macOS**, packed binaries are killed by the kernel)
3. TinyGo
4. shared-library or plugin boundary redesign

These are not default recommendations. UPX on Linux can achieve ~50% additional on-disk compression but increases startup time and RSS.

## Stop Conditions

Stop when one of these is true:

1. The next reduction would weaken runtime behavior or observability too much.
2. The remaining artifact size is dominated by required dependencies or assets.
3. The requested release constraints forbid stronger shrink tactics.
4. The binary is already small enough relative to the delivery mechanism.
