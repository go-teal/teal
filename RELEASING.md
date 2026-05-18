# Releasing

Teal is published as both a Go library (`github.com/go-teal/teal`) and a CLI binary (`teal`). A single git tag triggers both.

## How to cut a release

1. Make sure `main` is green and you have pulled the latest:
   ```bash
   git checkout main
   git pull
   ```
2. Pick the next semver tag (`vMAJOR.MINOR.PATCH`).
3. Tag and push:
   ```bash
   git tag -a v0.3.0 -m "v0.3.0"
   git push origin v0.3.0
   ```
4. The `Release` workflow in `.github/workflows/release.yml` will:
   - Build CLI binaries for linux / macOS / windows × amd64 / arm64
   - Inject the tag into `pkg/configs.TEAL_VERSION` via `-ldflags`
   - Generate a changelog from commits since the previous tag (grouped by type)
   - Publish a GitHub Release with binaries, checksums, and the changelog

That's it. No manual builds, no hand-written changelog.

## Library consumers

Library users just do `go get github.com/go-teal/teal@v0.3.0` — Go's module proxy picks up the new tag automatically. No extra step needed.

## Commit message conventions

The changelog groups commits by prefix. Use these to get clean release notes:

| Prefix          | Section              |
| --------------- | -------------------- |
| `feat:`         | Features             |
| `fix:`          | Bug fixes            |
| `sec:` / `security:` | Security        |
| `docs:`         | Documentation        |
| `chore(deps):` / `build(deps):` | Dependency updates |
| anything else   | Other                |
| `test:` / `ci:` | excluded             |

Scope is optional: `feat(cli): add --json flag`. Breaking changes: add `!` (`feat!: ...`).

## Pre-releases

Tag with a suffix to publish a pre-release: `v0.3.0-rc.1`. GoReleaser marks it as a pre-release automatically (`prerelease: auto`).

## Local dry-run

Before pushing a tag, you can preview what GoReleaser will produce:

```bash
goreleaser release --snapshot --clean --skip=publish
```

Artifacts land in `./dist/`. Requires [GoReleaser](https://goreleaser.com/install/) installed locally.

## Version visibility

`teal version` prints the value of `pkg/configs.TEAL_VERSION`. On released binaries this is the git tag (e.g. `v0.3.0`); on a `go build` from source it falls back to `dev`.
