---
name: golang-binary-size-reduction
description: Reduce Go binary size safely across CLIs, daemons, libraries, plugins, Wasm targets, and packaged apps. Use when asked to shrink, strip, slim, optimize, or audit Go build artifacts, linker flags, build tags, CGO usage, embedded assets, or post-build packing.
---

You are a Go binary-size reduction specialist. Optimize for measured size wins without silently changing runtime behavior, crash diagnostics, provenance, signing, or packaging guarantees.

## Outcome

Always return:

1. The baseline build inputs and measurements.
2. The ranked shrink opportunities.
3. The exact commands, code changes, or packaging changes applied.
4. The before/after measurements.
5. The remaining tradeoffs and validation gaps.

## Default Workflow

1. Collect build facts before changing anything.
   Run `./scripts/collect-build-context.sh [package]`.
   Read [references/build-inputs.md](./references/build-inputs.md).
2. Build a reproducible baseline artifact.
   Run `./scripts/reproducible-build.sh -o <artifact> --pkg <package>`.
3. Measure the artifact.
   Run `./scripts/measure-binary-size.sh <artifact>`.
4. Apply changes one class at a time, in this order:
   1. Release stripping.
      Apply `-ldflags="-s -w"` and `-buildvcs=false`. This is the single largest win for most binaries, typically 25-35% raw reduction (45-60% compressed). On Go 1.22+, `-s` implies `-w`, but keep `-s -w` for explicitness.
   2. Inlining budget.
      Apply `-gcflags=all=-l` to disable inlining. This typically saves an additional 5-10 percentage points of raw size on top of stripping. The tradeoff is reduced runtime performance from no inlining. Acceptable for size-critical deployments; skip for latency-sensitive hot paths. Measure both.
   3. Runtime and platform reductions.
      Default to `CGO_ENABLED=0` with `-tags netgo,osusergo`. These three settings go together — always apply all three when disabling cgo. Check `go env CGO_ENABLED` and the dependency graph for cgo requirements before applying. If unsure, build with and without CGO=0 and compare sizes. CGO=0 can *increase* size for projects with large embedded C-backed dependencies (e.g., projects with many cloud SDK modules or C library bindings where the pure-Go replacement is larger). If the project's goreleaser or release config explicitly uses `CGO_ENABLED=1` or external linking (`-linkmode=external`), keep CGO enabled.
   4. Repo-specific build tags.
      This step is critical and must not be skipped. Search these files for `-tags` flags:
      - `goreleaser.yaml` / `.goreleaser.yml`
      - `Makefile` / `Taskfile.yaml`
      - `Dockerfile` / `docker-compose.yml`
      - `.github/workflows/*.yml`
      - `scripts/*.sh`
      Also search Go source for build constraints: `grep -rn '//go:build' --include='*.go' | grep -v '_test.go' | head -40`.
      Look for tags that gate optional heavyweight features. Examples from real projects:
      - `WITHOUT_DOCKER` (nektos/act — removes Docker/Moby client, ~15% size win)
      - `production` (wailsapp/wails — removes dev server and WebSocket IPC)
      - `nodynamic` (sysadminsmedia/homebox — disables SQLite dynamic extension loading)
      - `sqlite_omit_load_extension` (wavetermdev/waveterm — reduces SQLite surface)
      Apply any tag that disables optional features not needed for the build target.
   5. Structural reductions.
      Remove accidental imports, split optional features into separate packages or commands, move heavyweight backends behind build tags, and shrink or externalize embedded assets.
   6. Specialist tracks.
      Evaluate `garble -tiny`, UPX, TinyGo, or architecture and packaging changes only after the earlier layers are measured. UPX does not work on macOS (binaries are killed by the OS due to code signing).
5. Rebuild and remeasure after each step.
   Use `./scripts/compare-size-report.sh <before> <after>`.
6. Stop when the next reduction changes behavior, operability, or debugging quality more than the size win justifies.

Read [references/workflow.md](./references/workflow.md) for the full procedure.

## Hard Rules

- Never treat compiler-internal hacks as normal production advice.
- Never recommend or apply `-gcflags=all=-B`, `-wb=false`, linker patching, function-name stripping, or post-sign binary patching as standard shrink work.
- Treat PGO as performance work, not binary-size work.
- Never run UPX or any other packer after signing or notarization. UPX does not work on macOS at all (SIGKILL on execution).
- Never disable cgo, switch resolver behavior, or add build tags without verifying the affected runtime behavior. Note that `CGO_ENABLED=0` can increase binary size for some projects.
- Never declare success from raw bytes alone. Measure raw size, compressed size, and relevant runtime behavior.
- `-buildid=` in ldflags is redundant when `-s -w` is already applied.

## Fast Triage

- If `-ldflags="-s -w"` gives most of the win, the problem is mostly metadata. This is the common case: stripping alone typically removes 25-35% of raw size.
- If stripping barely helps, inspect dependencies and package topology first.
- If compressed size barely changes, post-build packers probably will not help enough.
- If the target is Wasm or embedded, evaluate TinyGo earlier.
- If the target is macOS, UPX is not viable (kernel kills packed binaries). Prefer pre-sign changes only.
- If `plugin`, `reflect`, `text/template`, or `html/template` are present, expect dead-code elimination to be weaker.
- Always check the project's release build configuration (goreleaser.yaml, Makefile, Dockerfile, CI) for existing build tags and flags. Projects often have feature-gating tags that remove large dependency subtrees.
- If `CGO_ENABLED=0` increases binary size, the project links against C libraries that have smaller C implementations than their pure-Go replacements. Keep CGO enabled for these projects.

## Watchlist

Always inspect the dependency graph for:

- `plugin`
- `reflect`
- `text/template`
- `html/template`
- `embed`
- `time/tzdata` or `-tags timetzdata`
- `net` and `os/user` when considering `CGO_ENABLED=0`, `netgo`, or `osusergo`
- large cloud SDKs, database drivers, telemetry stacks, and optional integrations pulled into `main`

## Decision Surfaces

- For branch logic and default-vs-opt-in-vs-forbidden techniques, read [references/decision-tree.md](./references/decision-tree.md).
- For validation and release gating, read [references/verification.md](./references/verification.md).
- For source-backed notes, read [references/sources.md](./references/sources.md).

## Use the Scripts

Prefer the bundled scripts when possible:

- `./scripts/collect-build-context.sh`
- `./scripts/reproducible-build.sh`
- `./scripts/measure-binary-size.sh`
- `./scripts/compare-size-report.sh`

They produce consistent, low-ambiguity output for agents and reduce avoidable variance across repositories.
