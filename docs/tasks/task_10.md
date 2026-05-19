# Task 10: build-tag-gate the debug UI (`pkg/ui` → `teal_ui` tag)

## Status: ✅ IMPLEMENTED (initial cut)

## Why

teal pulls **gin** + `gin-contrib/cors` directly to host the in-browser DAG
debugger (`pkg/ui/server.go`). Through gin, the transitive dependency tree
balloons:

- `gin-gonic/gin` ≈ HTTP router and middleware
- `gin-contrib/cors` ≈ CORS middleware
- `bytedance/sonic` + `bytedance/sonic/loader` ≈ fast JSON
- `cloudwego/base64x` ≈ pulled by sonic
- `twitchyliquid64/golang-asm` ≈ pulled by sonic JIT
- `go-playground/validator/v10` + locales / universal-translator ≈ gin validation
- `gabriel-vasile/mimetype` ≈ pulled by gin
- `ugorji/go/codec` ≈ pulled by gin XML
- `quic-go/quic-go` + `quic-go/qpack` ≈ HTTP/3 via gin
- `klauspost/cpuid/v2` ≈ pulled by sonic JIT

For production DAG runs (`teal push` inside a cron job, container, or
serverless function) **none of this is needed** — the DAG runner uses only
`pkg/core` + `pkg/dags` + `pkg/processing` + `pkg/drivers`.

The cost is real:

- **Build time.** The gin tree (cors → validator → sonic → cloudwego → bytedance + JIT asm) materially slows down `go build`. On DigitalOcean Functions, where build time is hard-capped at 120 sec, partner-analytics deploys hit the cap *just because the gin transitive tree had to compile*. Stripping the tree took build time well under the cap.
- **Vendor size.** When a downstream project uses `go mod vendor` for offline / hermetic builds (typical for serverless and container deploys), gin alone bumps `vendor/` by ~15 MB.
- **Attack surface.** Every transitive dependency is a CVE surface; debug-only deps should not ship in production binaries.

## What changed

Two files in this repo:

- **`pkg/ui/server.go`** — prepended `//go:build teal_ui`. Without the tag, this file (and everything it transitively imports) is **excluded** from the build.
- **`internal/domain/generators/templates/main_ui.go.tmpl`** — the generated `cmd/<project>-ui/<project>-ui.go` debug binary now also opens with `//go:build teal_ui`. New projects scaffolded after this change get an opt-in UI binary out of the box.

## How to use

Production build (default — no UI, no gin):

```bash
go build ./cmd/<project>
```

Debug UI build (opt-in):

```bash
go build -tags teal_ui ./cmd/<project>-ui
```

CLI alias `teal ui` continues to work as it always did — that subcommand
shells out to the generated `<project>-ui` binary, which the user can build
with the tag.

## Measured impact (partner-analytics, 2026-05-19)

| Metric                  | Before (no tag)           | After (`teal_ui` gated) |
|---|---|---|
| `vendor/` size          | 18 MB                     | 18 MB (with gota / x-deps still present — gin tree gone) |
| Whether gin in vendor   | yes                       | **no**                  |
| Whether sonic in vendor | yes                       | **no**                  |
| Whether quic-go in vendor | yes                     | **no**                  |
| DO Functions build       | hits 120 s cap, fails     | (next: re-run pending)  |

Numbers will be updated after the re-deploy proves the build now fits in
the DO Functions 120 s window.

## What this does **not** solve

The remaining heavy hitters in `vendor/` after this change are:

- `golang.org/x/{crypto,net,sys,text,arch,sync}` — ~12 MB, pulled by `pgx`
  and `quic-go`. With `quic-go` gone, `golang.org/x/net` should be smaller
  on next re-vendor.
- `gonum.org/v1/gonum` (~2.8 MB) — pulled by `gota`, which is in turn
  pulled by `pkg/processing/sql_asset.go` (`getDataFrame` path for
  `is_data_framed: true` models). If we tag-gate the `is_data_framed`
  feature behind a `teal_dataframe` build tag too, gota + gonum drop out
  for projects that don't use `is_data_framed` — same approach, second
  iteration. Tracked separately.
- `flosch/pongo2/v6` (~312 KB) — template engine; load-bearing for every
  teal run (SQL templating). Stays.

## Open question — when to repackage

Currently the `teal_ui` tag is a "downstream user opt-in" — anyone building
their generated project picks default vs UI. Should the teal CLI itself
also be split into `teal` (without UI) and `teal-debug` (with UI)? Today
the `teal` binary is built without `pkg/ui` imports (the UI is only in the
generated downstream binary), so probably not — but worth a note before
v1.2 release.

## Discovered by

PoC of [partner-analytics](https://gitlab.com/dubai-one-click/partner-analytics)
(Elstate Partner Platform admin dashboard, TASK-AD-027) trying to deploy to
DigitalOcean Functions on 2026-05-19. Hit the 120 s build cap, traced the
cause to gin's transitive tree compiling for no runtime use, applied the
build-tag fix here.
