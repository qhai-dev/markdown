# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Single-binary Go port of [mdBook](https://github.com/rust-lang/mdBook). Reads `devdoc.yaml` plus per-file YAML front-matter (`title:` required, `index:` optional) and emits rendered HTML. Internal layout mirrors the Rust original; see `docs/old-plan/` for the original rewrite plan and historical status, and `docs/` for current architecture notes (configuration, crate mapping, runner vs Rust).

Go version is pinned in `go.mod` (currently `go 1.26.4`). Binary path is `bin/devdoc` (or `bin/devdoc.exe` on Windows).

## Build / test / run

The Makefile is the source of truth. Targets:

```bash
make build          # bin/devdoc(.exe) with -trimpath, stripped, version injected
make build VERSION=v1.0.0   # override the -X injected version
make dev            # debug build: keeps symbols + paths (for dlv / stack traces)
make version        # build + print version (sanity-check the -X flag)
make test           # go test ./...
make serve          # build + serve tests/basic on :3000 (BOOK=/path PORT=port to override)
make clean          # rm bin/devdoc
```

Direct Go incantations for ad-hoc work:

```bash
go build -o bin/devdoc ./cmd/devdoc
./bin/devdoc build --dir tests/basic --dest-dir /tmp/out
./bin/devdoc init  [--dir DIR] [--theme]   # scaffold a new book
go test ./...                              # full test suite (also runs *_test.go files under tests/*)
go test ./internal/runner/...             # one package
go test -run TestName ./pkg/foo            # one test
```

Subcommands are wired in `pkg/cli/cli.go` (cobra): `build`, `clean`, `init`, `run` (serve + open), `version`, `watch`. New subcommands must be registered in `New()` there.

## Strict-mode Rust vs Go diff harness

```bash
./harness/diff.sh [fixture ...]            # default: all tests/* fixtures
MDBOOK_RUST_BIN=/path/to/mdbook ./harness/diff.sh basic nested
```

Builds the Rust `mdbook` (debug) and the Go binary, runs both on the same fixture into temp dirs, then `diff -r`s. Since M2 the comparison is strict; intentional deviations live in `harness/KNOWN_DIFFS.md`.

Caveat: a deliberate front-end fork (writer.css, outline rail, dropped hljs/github-markdown switch) means byte-level diff is currently disabled by default. Set `MDBOOK_NO_FRONTEND_DIFF=0` to re-enable per-fixture byte diff, or rely on the markdown golden tests under `go test ./...` for content-level regression.

The harness also auto-skips `external-plugin` (M3 external-plugin link is frozen — see `internal/plugin/cmd.go` FROZEN note and `doc/plan/progress.md`) and `ts-markdown-basic_markdown` (goldmark-vs-pulldown-cmark HTML block boundary, in `MIGRATION.md`).

## High-level architecture

The binary is structured to mirror mdBook's crate split. Read top-down:

- **`cmd/devdoc/main.go`** — entry point. Single hard-coded `exitCode = 101` on error (mirrors Rust `std::process::exit(101)`); `formatError` walks the `errors.Unwrap` chain into Rust-style "Caused by:" lines.
- **`pkg/cli/`** — cobra commands. `cli.go` registers subcommands; each subcommand has its own file (`build.go`, `serve.go` lives in `run.go`, `watch.go`, etc.). Subcommands call into `internal/runner`.
- **`pkg/fs/`** — small path + file-copy helpers. Used by the runner.
- **`internal/model/`** — `Book`, `Chapter`, `SectionNumber` data model with hierarchical numbering; `devdoc.yaml` parser; HTML config. This is the typed spine.
- **`internal/runner/`** — orchestrator. Owns the loader, build pipeline, init flow, builtin preprocessors (links/index/cmd), and the preprocessor registry. New preprocessors plug in here.
- **`internal/html/`** — `goldmark` Markdown AST → mdBook-shaped node tree → HTML serialization. Includes per-extension support: tables, footnotes, task lists, strikethrough, definition lists, admonitions, smart punctuation, math, hide-lines, font-awesome. Also does title IDs/dedup and `.md`→`.html` link rewriting.
- **`internal/html_template/`** — Handlebars subset engine (`internal/hbs/`), SHA-256 asset fingerprinting, `{{ resource }}` rewrite, theme engine, `elasticlunr`-compatible `searchindex.js` (Porter stemmer + stop words).
- **`internal/frontend/`** — `go:embed`'d default theme (`assets/` + `index.html`). User `theme/` overrides file-by-file.
- **`internal/plugin/`** — external preprocessor / renderer wire protocol. M3 status: code is in place but the **external-plugin link is frozen** — see the FROZEN comment at the top of `internal/plugin/cmd.go` and `doc/plan/progress.md`. Do not extend external-plugin acceptance without un-freezing first.

Per-chapter subdirectory preservation, prefix/numbered/suffix zones, part titles, separators, draft chapters, redirects, `additional-css`, `fold`, and per-file front-matter (with UTF-8 BOM stripping) are handled in the loader/walk + model layer.

## Test fixtures

- `tests/basic` — minimal book used by `make serve` and the harness.
- `tests/nested` — multi-level nesting, tables, footnotes, admonitions, task lists, additional-css, fold, draft chapters, part titles, separators, prefix/numbered/suffix zones, redirects. Used as a strict-mode harness fixture.
- `examples/basic` — example output, also serve-able via `make example`.
- `tests/ts-*` — TypeScript-style fixtures historically imported from the Rust testsuite (see `tests/README.md`).
- `bazel/tests/` — Bazel build fixtures, orthogonal to the Go harness.

## Roadmap (from `README.md`)

Closed: M2 HTML renderer (basic + nested byte-identical to Rust build under strict mode).

Open milestones:
- **M3** — preprocessor / renderer plugin protocol (wire in `internal/plugin`; builtin impls in `internal/runner`). External-plugin acceptance is **frozen**.
- **M4** — `test`, `clean` subcommands.
- **M5** — `watch`, `serve`, live reload (`serve` already lands; `watch` is wired but not exercised).
- **M6** — regression matrix, cross-platform builds, perf benchmarks.

Before touching milestone plumbing, skim `docs/old-plan/progress.md` (if present in your checkout) and the FROZEN notes in the relevant files.