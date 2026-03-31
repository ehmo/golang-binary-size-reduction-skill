# Go Binary Size Reduction Skill

Shrink Go binaries by 30-45% without breaking anything. Works with Claude Code, OpenAI Codex, and any agent that reads skill files.

Your agent collects build facts first, applies safe reductions in order of impact, measures after each change, and stops before breaking runtime behavior. No folklore flags, no cargo-culted compiler hacks.

## What it does

Your agent gets a structured workflow:

1. Collect build context (Go version, CGO state, build tags, dependencies).
2. Build and measure a reproducible baseline.
3. Apply reductions one class at a time, measuring after each step.
4. Report the before/after numbers and the tradeoffs.

It includes shell scripts for reproducible measurement, a decision tree for edge cases, and hard rules that prevent the agent from recommending unsafe techniques.

## Benchmark results

I tested this skill against 14 trending Go repositories from GitHub. Each repo was pinned to a specific commit, built with default flags as a baseline, then optimized by an agent using the skill. All builds passed smoke checks (the binary starts, prints help, responds to basic commands).

| Repository | Baseline | Raw reduction | gzip reduction | xz reduction |
|---|---|---|---|---|
| nektos/act | 28.0 MB | 45.9% | 61.9% | 67.5% |
| wavetermdev/waveterm | 35.4 MB | 41.6% | 60.5% | 67.4% |
| danielmiessler/Fabric | 88.0 MB | 36.9% | 56.7% | 65.9% |
| XTLS/Xray-core | 41.6 MB | 39.3% | 57.9% | 65.3% |
| henrygd/beszel | 12.1 MB | 37.2% | 56.0% | 62.4% |
| masterking32/MasterDnsVPN | 5.7 MB | 35.4% | 55.2% | 61.4% |
| mostlygeek/llama-swap | 8.8 MB | 37.7% | 55.7% | 61.7% |
| usememos/memos | 51.9 MB | 41.1% | 59.7% | 66.7% |
| sysadminsmedia/homebox | 125.4 MB | 35.5% | 53.7% | 61.6% |
| authelia/authelia | 63.4 MB | 39.2% | 58.6% | 65.4% |
| wailsapp/wails | 31.3 MB | 31.1% | 45.2% | 51.9% |
| owncast/owncast | 96.3 MB | 24.3% | 39.7% | 46.3% |
| sundowndev/phoneinfoga | 43.6 MB | 30.2% | 46.7% | 56.0% |
| supabase/cli | 124.3 MB | 37.0% | 58.2% | 66.5% |

**Average raw reduction: 36.6%.** Average gzip reduction: 54.7%. Average xz reduction: 61.8%.

All 14 repos passed build and smoke checks. No binary failed to start.

Reductions came from stripping debug symbols (`-ldflags="-s -w"`), disabling inlining (`-gcflags=all=-l`), toggling CGO where safe, and discovering project-specific build tags that remove optional dependency subtrees.

## How it was built

The skill started as a collection of notes from the Go toolchain docs, production case studies (Datadog, Tailscale), and practical blog posts. I organized those into a decision tree and a set of hard rules about what to never do (disable bounds checks, UPX on macOS, post-sign patching).

Then I tested it. I wrote a benchmark harness that snapshots the GitHub Go trending page, pins each repo by commit SHA, builds deterministic baselines, and hands the agent a task: shrink this binary, report the numbers. The harness measures raw size, gzip size, xz size, startup time, and RSS. It runs smoke checks to confirm the binary still works.

I iterated on the skill through 13 benchmark rounds. Each round revealed gaps: projects where the agent missed feature-gating build tags, repos where CGO_ENABLED=0 made the binary bigger, edge cases around macOS signing. I patched the skill after each round and re-ran. The version in this repo is the result of that loop.

The benchmark harness lives in a separate development repo. This repo contains the tested, final skill.

## Installation

```bash
npx skills add ehmo/golang-binary-size-reduction-skill
```

Or clone this repo and point your agent at it.

## Usage

Invoke it directly:

```
Shrink the binary for ./cmd/myapp and report the before/after sizes.
```

```
Audit this Go project for binary size reduction opportunities.
```

```
Apply release stripping and measure the impact.
```

Four bundled shell scripts produce consistent output across repos:

- `scripts/collect-build-context.sh` -- gathers Go version, CGO state, dependencies, and watchlist hits
- `scripts/reproducible-build.sh` -- builds with stable flags and paths
- `scripts/measure-binary-size.sh` -- reports raw, gzip, and xz sizes plus top symbols
- `scripts/compare-size-report.sh` -- diffs two artifacts with percentage changes

## Repo structure

```
SKILL.md              -- Agent instructions with YAML frontmatter
AGENTS.md             -- Quick context for agent consumption
metadata.json         -- Version, references, abstract
references/
  build-inputs.md     -- Facts to collect before starting
  decision-tree.md    -- What to try, what to skip, what's forbidden
  workflow.md         -- Step-by-step procedure
  verification.md    -- How to validate the result
  sources.md          -- Authoritative references
scripts/
  collect-build-context.sh
  reproducible-build.sh
  measure-binary-size.sh
  compare-size-report.sh
agents/
  openai.yaml         -- OpenAI Codex agent config
```

## Agents

Works with any agent that reads skill files. Tested with:

- **Claude Code** -- reads `SKILL.md` directly
- **OpenAI Codex** -- uses the `agents/openai.yaml` config

## Sources

- Go build, linker, and embed docs (pkg.go.dev)
- Datadog: shrinking agent binaries (datadoghq.com/blog)
- Tailscale: small binary docs and macOS binary size work (tailscale.com)
- Filippo Valsorda, Liam Stanley, Garble, UPX, TinyGo docs

Full list in `references/sources.md`.

## Contributing

- Keep PRs focused to one change.
- If updating rules or decision logic, explain why and include a before/after example.
- Do not add techniques to the decision tree without testing them against at least a few real repos.

## License

MIT
