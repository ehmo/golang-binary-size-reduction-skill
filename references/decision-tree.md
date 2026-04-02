# Decision Tree

Use this file to decide what to try next.

## Default Techniques

These are the safest first moves.

| Technique | Use when | Notes |
| --- | --- | --- |
| `-trimpath` | almost always | Good hygiene and reproducibility; small size win at best |
| dependency pruning | almost always | Highest-value durable reductions |
| build-tag gating | optional features exist | Best when optional backends or cloud integrations are present |
| reduce or externalize embedded assets | `embed` is present | Binary size often follows asset size directly |
| avoid `timetzdata` | timezone DB is not required in-binary | `time/tzdata` adds about 450 KB |

## Opt-In Techniques

Use only after confirming the tradeoff is acceptable.

| Technique | Use when | Typical raw win | Risk |
| --- | --- | --- | --- |
| `-ldflags="-s -w"` | release artifact can lose DWARF and symbol table data | 25-35% | weaker debugging and symbolization |
| `-buildvcs=false` | provenance is captured elsewhere | <1% | weaker embedded provenance |
| `-gcflags=all=-l` | performance regression from no inlining is acceptable | 5-10% additional | reduced runtime performance |
| `CGO_ENABLED=0` | cgo is not required | 0-6% (varies) | behavior changes; can *increase* size for some projects |
| `CGO_CFLAGS="-Oz"` | cgo must stay enabled and C source is compiled | 1-2% | slightly slower C code execution; zero effect if CGO only links system libraries |
| `-tags netgo` | pure-Go DNS behavior is acceptable | <1% | resolver differences |
| `-tags osusergo` | pure-Go user/group lookup is acceptable | <1% | user lookup differences |
| project-specific tags | project has feature-gating tags (check Makefile/goreleaser) | 0-15% | feature removal |
| `garble -tiny` | obfuscation is acceptable and crash output can be weaker | varies | harder debugging and crash analysis |
| UPX | Linux/Windows only; distribution model tolerates packers | ~50% on-disk | startup, memory, AV, signing; **does not work on macOS** |
| TinyGo | alternate compiler/runtime is acceptable | varies | compatibility and runtime differences |

## Forbidden Techniques

Do not present these as standard production advice.

| Technique | Status | Why |
| --- | --- | --- |
| `-gcflags=all=-B` | forbidden | disables bounds checks |
| `-gcflags=all=-wb=false` | forbidden | breaks GC invariants and is obsolete |
| linker patching or function-name stripping hacks | forbidden | unsupported and fragile |
| post-sign packing or patching | forbidden | breaks signatures and notarization |
| UPX on macOS | forbidden | packed binaries are killed by the kernel (SIGKILL) |
| PGO as a size tactic | do not use | can increase binary size |
| `-compressdwarf=false` as a shrink step | do not use | increases size; only for debugger compatibility |
| `-ldflags="-buildid="` with `-s -w` | redundant | build ID is already stripped by `-s -w`; adds no additional savings |
| `CC="zig cc"` as a size technique | do not use | zig cc does not reduce binary size; benchmarks show 0-0.8% *increase* on native builds; zig is a cross-compilation tool, not a size tool |

## Repo-Specific Tag Discovery

Before applying generic flags, check the project's build infrastructure for existing feature-gating tags:

1. Read `goreleaser.yaml`, `Makefile`, `Taskfile.yaml`, `Dockerfile`, and CI workflow files.
2. Look for `-tags` in build commands. Projects often have tags like:
   - `WITHOUT_DOCKER` (nektos/act) — removes Docker/Moby client, ~15% additional reduction
   - `production` (wailsapp/wails) — removes dev server code
   - `nodynamic` (sysadminsmedia/homebox) — enables static SQLite build
   - `sqlite_omit_load_extension` (wavetermdev/waveterm) — reduces SQLite surface
3. Search source for `//go:build` and `// +build` constraints to find optional features.
4. These tags can remove entire dependency subtrees, producing larger wins than flag tuning.

## Branches

### If stripping helps a lot

The binary is metadata-heavy. Keep the change only if the artifact can lose debug data safely.

### If stripping barely helps

Focus on structure:

1. prune direct imports from `main`
2. split optional features
3. gate integrations with build tags
4. reduce embedded assets
5. inspect `plugin`, `reflect`, and template packages

### If cgo is present

Ask whether it is actually required.

If yes:

- keep cgo on
- if the project compiles C source code (e.g., `mattn/go-sqlite3` amalgamation), apply `CGO_CFLAGS="-Oz"` for 1-2% raw savings
- focus on dependency and asset reduction

If no:

- measure a `CGO_ENABLED=0` build
- verify networking and user lookup behavior

### If `plugin` is present

Assume dead-code elimination is weaker. Try to isolate plugin support into a separate binary or build-tagged path.

### If `embed` or `time/tzdata` is present

Treat those bytes as intentional payload, not linker waste. Remove or move them only if the runtime contract allows it.

### If the artifact is already compressed in transit

Measure gzip or xz size before considering UPX. Packers help less when the binary is already distributed inside a compressed medium.

### If the target is macOS and signed

Only make size changes before signing. Re-run signing and notarization validation after the final build.
