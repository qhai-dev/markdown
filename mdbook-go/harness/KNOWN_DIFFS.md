# Known Rust vs Go differences

The diff harness exits non-zero on any byte difference between the Rust
and Go outputs. Every line below is **expected** and corresponds to either
an intentionally skipped fixture or a parser-level deviation that requires
non-trivial work to close.

Last updated: 2026-08-09.

## Front-end fork (Writer 观感) — ALL fixtures

2026-08-09 起前端有意与 Rust 端分叉，逐字节 diff 对所有 fixture 失效
（`diff.sh` / `diff_rust_testsuite.sh` 默认全部跳过，
`MDBOOK_NO_FRONTEND_DIFF=0` 可恢复）：

- 新增 `internal/static/css/writer.css`（Go-only）：设计令牌（主题 class
  映射的 color-mix 推导）+ `.markdown-body` 正文样式 + `.hljs-*` oklch
  色板 + 右侧大纲栏（`.rail-*`）。替代原 github-markdown/hljs 主题。
- 新增 `internal/static/js/outline-rail.js`（Go-only）：大纲栏交互
  （hover 弹卡片、点击跳转、滚动高亮、右键复制标题链接、Esc）。
- `templates/index.html`：head 换 writer.css 链接、加 rail-zone 标记与
  outline-rail.js 脚本。
- `js/book.js`：`themes()` 删掉 github-markdown/hljs 样式表切换
  （这些 css 仍随构建输出，作为备份留在仓库；恢复旧观感见
  templates/index.html 里的注释）。
- 参考 demo：`mdbook-go/test-html-css/index.html`。

内容级回归不再由 harness 承担，由 `go test ./...` 的 markdown golden
tests 覆盖。

## Skipped fixtures

(None. Both `basic` and `nested` pass strict diff.)

## Markdown parser deviations (goldmark vs pulldown-cmark)

These are tracked in `internal/html/markdown_golden_test.go`'s
`knownDeviations` slice. The corresponding fixtures under
`tests/testsuite/markdown/` are skipped from the golden regression. They
do **not** affect `basic` or `nested`; they only show up when reusing the
Rust `testsuite` fixtures as Go goldens, which is a separate test path.

1. `tests/testsuite/markdown/definition_lists/definition_lists.md` —
   goldmark requires a single-line plain-text term; inline links or
   multi-line terms do not become `<dt>`.
2. `tests/testsuite/markdown/basic_markdown/html.md` — when an opening
   HTML tag spans two lines, goldmark treats it as a block element while
   pulldown-cmark falls back to inline HTML inside a paragraph.

Fixing either requires swapping out part of goldmark's block parser;
deferred until a fixture explicitly demands it.

## Items removed from this file

These were listed under M1/M2 and have since been closed:

- `404.html`, `print.html`, `toc.html`, `toc.js`, `searchindex.js` — M2.
- Font Awesome CSS/JS/icons and the menu bar / sidebar / footer JS — M2.
- Theme asset hashing and `{{ resource }}` rewriting — M2.
- `redirect` table support — M2.
- `additional-css` and `fold` rendering — M2 (nested fixture).
- Strict-mode byte-for-byte equivalence on `basic` and `nested` — M2.