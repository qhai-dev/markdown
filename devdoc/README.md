# mdbook-go (devdoc)

Go port that replaces [mdBook](https://github.com/rust-lang/mdBook).
Everything under this directory is a single-binary Go implementation
that reads `devdoc.yaml` + per-file YAML front-matter (`title:` required,
`index:` optional) and emits the rendered HTML. See
[`docs/old-plan/`](docs/old-plan/) for the original rewrite plan and
historical status.

## Quick start

```bash
cd mdbook-go
go build -o bin/devdoc ./cmd/devdoc
./bin/devdoc build --dir tests/basic --dest-dir /tmp/out
./bin/devdoc init [--dir DIR] [--theme]   # scaffold a new book
```

## Layout


## Layout

```text
mdbook-go/
├── cmd/devdoc/       CLI entry point (build, init, version)
├── internal/
│   ├── model/         book model + config (devdoc.yaml loading, html config)
│   ├── runner/        MDBook orchestrator: loader, build, init + builtin
│   │                  preprocessors (links/index/cmd) + preprocessor registry
│   ├── frontend/      go:embed default frontend assets: assets/ (CSS+JS) + index.html (原 theme，2026-08-18 由 static/ 改名)
│   ├── hbs/           Handlebars subset engine (standalone whitespace, helpers)
│   ├── html/          goldmark → mdBook node tree → HTML serialization
│   ├── html_template/ static asset hashing + {{ resource }} rewrite, theme engine
│   ├── search/        elasticlunr-compatible searchindex.js
│   ├── plugin/        external preprocessor / renderer protocol (interfaces + wire)
├── pkg/
│   ├── cmd/           CLI subcommands (build, init, clean, open, serve, watch)
│   └── fs/            path helpers + file copy/write utilities
├── tests/          shared test books (basic, nested)
├── harness/           Rust-vs-Go diff harness (strict mode)
```

## Current milestone

**M2: HTML renderer — closed. Strict-mode harness passes on basic + nested.**

Verified:

- `tests/basic` — 40 files byte-identical to the Rust build.
- `tests/nested` — 48 files byte-identical to the Rust build
  (multi-level nesting, tables, footnotes, admonitions, task lists,
  `additional-css`, `fold`, draft chapters, part titles, separators,
  prefix / numbered / suffix zones, redirects).
- `internal/html` golden tests pass against the Rust test suite fixtures.

Implemented:

- `Book`, `Chapter`, `SectionNumber` data model with hierarchical
  numbering and per-chapter subdirectory preservation.
- `devdoc.yaml` (YAML) parsing with dynamic `output.*` and
  `preprocessor.*` sections.
- `[chapters]` config (replacing `SUMMARY.md`): arbitrary nesting,
  prefix / numbered / suffix zones, part title, separator, draft,
  subdirectory chapters.
- Disk loader with UTF-8 BOM stripping.
- `goldmark` driven Markdown → mdBook node tree → HTML, with extensions
  for tables, footnotes, task lists, strikethrough, definition lists,
  admonitions, smart punctuation, math, hide-lines, font awesome.
- Title IDs and dedup; `.md` → `.html` link rewriting.
- `index.html`, per-chapter pages, `toc.html`, `toc.js`,
  `404.html`, redirects, `.nojekyll`.
- Static asset collection, SHA-256 fingerprinting, `{{ resource }}`
  rewriting (CSS/JS/font/icons).
- Inlined default theme via `go:embed`; user `theme/` overrides file by
  file.
- `elasticlunr`-compatible `searchindex.js` (Porter stemmer + stop
  words) — landed early because the chapter `<head>` references its
  hashed name.
- `init` and `build` subcommands.

Not yet implemented:

- M3: preprocessor / renderer plugin protocol (wire protocol in `internal/plugin`; builtin impls in `internal/runner`; external-plugin acceptance **frozen**, see cmd.go FROZEN note)
- M4: `test`, `clean` subcommands
- M5: `watch`, `serve`, live reload
- M6: regression matrix, cross-platform builds, performance benchmarks

## Harness

```bash
./harness/diff.sh [fixture ...]
```

The harness builds both the Rust and Go binaries, runs `mdbook build` on
the same fixture into separate output directories, then `diff -r`s the
result.

Since M2 the comparison is strict: any difference is a failure. The only
known allowed deviations live in `harness/KNOWN_DIFFS.md` and currently
cover only two goldmark vs pulldown-cmark parser quirks — neither of
which is exercised by `basic` or `nested`.

```bash
# Override the Rust binary location if needed:
MDBOOK_RUST_BIN=/path/to/mdbook ./harness/diff.sh basic nested
```

## Running the Rust test suite

```bash
go test ./...
```

The Go side has its own fixture set under `tests/`; historically some
were imported from the Rust `tests/testsuite/` while that source was
available.