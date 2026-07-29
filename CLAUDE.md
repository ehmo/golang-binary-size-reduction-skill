# Go Binary Size Reduction Skill

Shrink Go binaries with measured tradeoffs.

- `SKILL.md` -- Full guidelines with YAML frontmatter
- `metadata.json` -- Version, references, abstract
- `analyzer/` -- Reusable `go/analysis` binary-size audit
- `references/` -- Decision tree, workflow, verification, build inputs, sources
- `scripts/` -- Shell scripts for source analysis and reproducible measurement
- `agents/` -- Agent configs (OpenAI Codex)

`collect-build-context.sh` runs the `binsize` analyzer over non-test files in `./...`.
