# Known harness differences

`harness/diff.sh` runs `devdoc build` against each fixture under
`tests/` and bytes-compares the generated `book/` against the
captured expected output. Every entry below is an **expected**
deviation: typically an intentionally-skipped fixture, an asset the
harness ignores, or a category of known mismatch the Go side won't
close.

## Skipped fixtures

(None at present. `basic` and `nested` both pass strict diff.)

## Front-end fork vs. historical baseline

All fixtures use Go-only front-end assets shipped under
`internal/frontend/`. These replace the historical CSS/JS theme stack
(fontAwesome / github-markdown-css / etc.). When generated outputs are
re-captured by running `devdoc build`, the `book/` is byte-stable
for the new stack.

The harness script itself does not compare against any external
reference; byte-equality is checked against the in-tree captured
output at `tests/<fixture>/book/`. To refresh expectations:
`devdoc build --dir tests/<fixture> --dest-dir tests/<fixture>/book`.

## Items removed from this file

Closed decisions; kept here as a project-history breadcrumb.

- `404.html`, `print.html`, `toc.html`, `toc.js`, `searchindex.js`,
  Font Awesome CSS/JS/icons, the menu-bar / sidebar / footer JS,
  Theme asset hashing, `{{ resource }}` rewriting — M2 removal.
- `redirect` table support — M2.
- `additional-css` and `fold` rendering — M2 (`nested` fixture).
- Strict-mode byte-for-byte equivalence on `basic` and `nested` — M2.
- **2026-08-15**: Left-side chapter list sidebar initially deleted,
  then partially restored. Today only the no-JS `toc.html` iframe
  fallback and the `SidebarHeaderNavSource` IIFE stay gone. The
  right-side outline rail (`outline-rail.js`) is the sole in-page
  outline.
- **2026-08-16**: `[chapters]` YAML config (`ChaptersConfig`,
  `ChapterItem`) removed entirely; chapter discovery is now
  filesystem-walk + per-file YAML front-matter (`title:` required,
  `index:` optional). See `docs/configuration.md` §4.
- **2026-08-16**: Section numbering (`Chapter.Number`, `SectionNumber`,
  `1.`/`1.1.` sidebar prefix) deleted. No replacement.
- **2026-08-16**: `Part` titles, `Separator` entries deleted. Nesting
  is expressed via filesystem subdirectory only.
- **2026-08-16**: All 18 fixtures' `devdoc.yaml` files stripped of
  `[chapters]` blocks; corresponding `tests/*/src/SUMMARY.md`
  Rust-leg files removed. Every `.md` carries `title:` front-matter.
- **2026-08-16**: Rust source tree (`crates/`, `src/`, `tests/testsuite/`,
  `tests/gui/`, `ci/`, `guide/`, `doc/`, root Rust configs) removed
  from the repo. The Go port `mdbook-go/` stands alone.
