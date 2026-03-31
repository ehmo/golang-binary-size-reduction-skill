# Sources

Use these sources to justify recommendations or resolve edge cases.

## Highest-Trust References

- Go build docs: https://pkg.go.dev/cmd/go
- Go build constraints: https://pkg.go.dev/cmd/go
- Go linker docs: https://go.dev/src/cmd/link/doc.go
- Go reproducible builds: https://go.dev/blog/rebuild
- Go PGO docs: https://go.dev/doc/pgo
- `debug/buildinfo`: https://pkg.go.dev/debug/buildinfo
- `os/user`: https://pkg.go.dev/os/user
- `net`: https://pkg.go.dev/net
- `embed`: https://pkg.go.dev/embed

## Strong Production Case Studies

- Datadog on shrinking agent binaries: https://www.datadoghq.com/blog/engineering/agent-go-binaries/
  Best source for dependency and package-boundary wins. Treat this as the main proof that structural changes beat folklore flags.
- Tailscale small-binary docs: https://tailscale.com/docs/how-to/set-up-small-tailscale
  Useful for build-tag strategy and for packer caveats.
- Tailscale macOS binary-size work: https://tailscale.com/blog/macos-binary-size
  Best source for macOS-specific signing and dead-code-strip constraints.

## Practical Tactic References

- Filippo Valsorda: https://words.filippo.io/shrink-your-go-binaries-with-this-one-weird-trick/
- Liam Stanley: https://liam.sh/p/shrinking-go-binaries
- Alexander Obregon: https://alexanderobregon.substack.com/p/go-binary-size-reduction
- Garble README: https://github.com/burrowers/garble
- UPX docs: https://upx.github.io/ and https://github.com/upx/upx
- TinyGo optimizing binaries: https://tinygo.org/docs/guides/optimizing-binaries/

## Historical or Experimental References

- xaionaro notes: https://github.com/xaionaro/documentation/blob/master/golang/reduce-binary-size.md
  Useful as a survey of old experiments, but do not treat removed or unsafe compiler flags as current advice.
- totallygamerjet smallest Go binary: https://totallygamerjet.hashnode.dev/the-smallest-go-binary-5kb
  Interesting for extreme-size experiments, not for general production guidance.
- Alexey Yuzhakov: https://sibprogrammer.medium.com/go-binary-optimization-tricks-648673cc64ac
  Useful for UPX measurements and tradeoffs.
- OneUptime: https://oneuptime.com/blog/post/2026-01-07-go-reduce-binary-size/view
  Useful as a current blog overview, but prefer primary sources for policy-level guidance.

## Supplied Skill Example

- samber Go linter skill: https://github.com/samber/cc-skills-golang/blob/main/skills/golang-linter/SKILL.md
  Useful as a style reference for a terse agent-facing skill layout.
