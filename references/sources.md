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

## CGO_CFLAGS and Zig cc Benchmarks

Benchmarked April 2026 on darwin/arm64 with Go 1.26.1 and Zig 0.15.2.

### CGO_CFLAGS="-Oz" (C source size optimization)

| Project | C dep | Default stripped | -Oz stripped | Raw delta |
|---------|-------|-----------------|--------------|-----------|
| owncast | mattn/go-sqlite3 | 79,635,282 | 79,125,266 | -0.64% |
| waveterm | mattn/go-sqlite3 | 24,111,858 | 23,601,826 | -2.11% |
| authelia | mattn/go-sqlite3 | 46,306,946 | 45,796,946 | -1.10% |
| wails | system frameworks | 24,599,282 | 24,599,282 | 0% |

Savings are consistent (~500 KB) across mattn/go-sqlite3 projects because the SQLite amalgamation is the same C source. Zero effect when CGO only links system libraries.

### CC="zig cc" (not a size technique)

| Project | Default CC | zig cc | Raw delta |
|---------|-----------|--------|-----------|
| act (pure Go) | 29,315,106 | 29,315,106 | 0% |
| owncast (sqlite3) | 79,635,282 | 79,797,794 | +0.20% |
| waveterm (sqlite3) | 24,111,858 | 24,299,570 | +0.78% |
| wails (system libs) | 24,599,282 | 24,599,282 | 0% |

Zig cc does not reduce binary size. On native macOS, it is clang underneath with slightly different defaults. Its value is cross-compilation convenience (single toolchain for all targets), not size.

- Zig cross-compilation overview: https://dev.to/kristoff/zig-makes-go-cross-compilation-just-work-29ho
- Uber's hermetic_cc_toolchain: https://github.com/uber/hermetic_cc_toolchain
- GoReleaser zig+cgo example: https://github.com/goreleaser/example-zig-cgo

## Supplied Skill Example

- samber Go linter skill: https://github.com/samber/cc-skills-golang/blob/main/skills/golang-linter/SKILL.md
  Useful as a style reference for a terse agent-facing skill layout.
