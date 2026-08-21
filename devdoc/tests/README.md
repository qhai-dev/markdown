# mdbook-go fixtures

This directory holds the books used by `devdoc build` smoke checks and
the `harness/diff.sh` regression script. Two categories exist today.

> 2026-08-16 起,目录树配置从 `[chapters]` YAML 块迁移到每文件
> front-matter (`title:` 必填,`index:` 可选)。fixture 中的 `SUMMARY.md`
> 已删除 — 嵌套通过子目录表达,章节顺序由 `index:` / 文件名控制。

## Hand-curated

| Name | What it covers |
|---|---|
| `basic/` | Single-chapter book with intro + two chapter pages, one sectioned page. Canonical smoke fixture. |
| `nested/` | Four-level nesting via subdirectories (`guide/`, `guide/advanced/`), tables, footnotes, admonitions, redirects. |
| `cli/` | Minimal book exercising `init` / `clean` / `build` CLI paths. |
| `serve/` | Book with `extra_watch_dirs`. |
| `external-plugin/` | Three preprocessor scripts (banner / footer / noisy); matches against an external plugin contract. |

## Bundled tests (`ts-*`)

Each `tests/ts-<category>-<name>/` is a self-contained book exercising a
specific feature (preprocessor directives, build options, redirects,
theme overrides, …). The `ts-` prefix lets `harness/diff.sh` match them
with a single glob (`tests/ts-*/`).

These were originally imported from the upstream `tests/testsuite/`
while the Rust sources were still in-tree (M6.1 era). After the Rust
sources were removed (2026-08-16), the fixtures stand on their own —
they contain `devdoc.yaml`, `src/`, and the expected `book/` output
captured at import time. To update a fixture's expected output, run
`/tmp/devdoc build --dir tests/ts-... --dest-dir tests/ts-.../book`.

## Adding a new fixture

1. Pick a focused scenario — one feature per `tests/ts-<feature>/` dir.
2. Write the book under `src/` with per-`.md front-matter
   (`title:` mandatory; `index:` optional for ordering).
3. Run `/tmp/devdoc build --dir tests/ts-<feature>` to populate
   `tests/ts-<feature>/book/` (the captured expected output).
4. If the fixture exposes a regression scenario `devdoc` should fix,
   add a row to this README explaining it.
