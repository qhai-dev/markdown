# Rust html_handlebars → Go 版分布盘点（聚合准备 + 历史归档）

> ⚠️ 已执行（2026-08-09）：本文描述的 7 包分散状态已被聚合为
> `internal/html_template` 单包（render/data/searchdocs/engine/
> helpers/static/theme/search/index/pipeline/stemmer + templates/）。
>
> ⚠️ 后续清理（2026-08-18）：移除 `[output.html.redirect]` 配置支持——
> `redirect.go` 整文件、`templates/redirect.html`、`Env.FragmentMap` 字段、
> `tests/ts-redirects-redirects_are_emitted_correctly/` 套件、`ts-redirects` 在
> `tests/nested/devdoc.yaml` 的用法均已删除。本文档下表中相关条目现已过期。
>
> ⚠️ 部分删除（2026-08-15）：章节列表 sidebar 渲染的"in-page 大纲 + noscript
> iframe 兜底"被删 —— `SidebarHeaderNavSource` IIFE、`toc.html` 模板、
> `Env.SidebarHeaderNav*` 字段、`index.html` 的 `<noscript><iframe>` 块。
> 仍然保留：`renderTocSidebar`、`Env.TocHTML`、`buildTocJS`、
> `MDBookSidebarScrollbox`、`<nav id="mdbook-sidebar">` 容器 + JS 注入
> 章节列表、`chapterSummaries`、`[output.html].fold.*` / `no-section-label` 配置。
> 见 `/Users/qhai-dev/.claude/plans/quirky-hopping-koala.md`。
>
> 更新时间：2026-08-09，分支 v1。
> 背景：Rust 的 HTML 渲染器是 `crates/mdbook-html/src/html_handlebars/`
> 一个模块（6 个 rs 文件 + 6 个主题模板 + 外部 handlebars crate）。
> Go 用 html/template 替代 handlebars（`internal/tplgotpl`），但这套逻辑
> 目前散落在 **7 个包**里。本文盘点每块逻辑现在在哪，为聚合做准备。

## 1. 分布总览（7 个包）

| Go 包 | 承载的 html_handlebars 部分 |
|---|---|
| `internal/render`（5 文件） | hbs_renderer.rs 本体 + 部分 search.rs + 模板注册 |
| `internal/tplgotpl`（4 文件 + prod/5 模板） | handlebars 引擎替代 + helpers 逻辑 + 模板文件 |
| `internal/frontend`（1 文件） | static_files.rs |
| `internal/search`（4 文件） | search.rs 的索引构建半 |
| `internal/theme`（1 文件） | theme/mod.rs 的资产解析半（模板不归它） |
| `internal/model`（html.go 一小块） | make_data 的 GitRepositoryIconClass |
| `internal/runner`（build.go 调用点） | Rust 的 `impl Renderer`（在 hbs_renderer.rs 内） |

## 2. 逐块分布表

### A. hbs_renderer.rs（696 行）→ 主要在 internal/render

| Rust 函数 | Go 位置 | 说明 |
|---|---|---|
| `HtmlHandlebars::render`（Renderer impl） | `render/render.go:43` `Render` | 编排主函数 |
| `render_chapter` | `render/render.go:332` `renderChapter` | |
| `render_404` | `render/render.go:402` `render404` | |
| `render_print_page` | — | print 功能已删 |
| 模板注册（index） | `render/render.go:156` `newRegistry` + `tplgotpl.LoadProduction` | handlebars.register_template_string → tplgotpl（redirect/head/header/toc_html 已删） |
| `register_hbs_helpers` | tplgotpl 的 Env 方法（TocHTML + Resource 等）；fa helper 已删 | |
| render "toc_js" 模板 | `render/render.go:182` `buildTocJS` + 3 个 JS 常量（含 `MDBookSidebarScrollbox` 注入章节列表 sidebar） | Rust 里是 theme/toc.hbs 模板，Go 硬编码。SidebarHeaderNavSource 已删（只保留章节列表 splice） |
| `make_data` | `render/data.go:85` `makeData`；另 `config/html.go:172` `GitRepositoryIconClass` | Go 另有 RenderData/BuildRenderData（类型化层，Go 独有） |
| `RenderChapterContext` | `render/render.go:27` `Context` + `renderChapter` 参数 | |

### B. helpers/（3 个 helper）→ tplgotpl + 零散

| Rust helper | Go 位置 | 散落情况 |
|---|---|---|
| `helpers/toc.rs::RenderToc`（{{#toc}}） | `tplgotpl/toc_render.go:15` `renderTocSidebar`（算法）<br>`tplgotpl/helpers.go:64` `Env.TocHTML`（方法）<br>`render/render.go:182` `buildTocJS`（toc.js 里的拼接） | **1 个 helper 拆了 3 处**（SidebarHeaderNavSource 拼接已删；`render/toc.go:NewSidebarEnv` 构造器也并入 buildTocJS） |
| `helpers/resources.rs::ResourceHelper`（{{resource}}） | `tplgotpl/helpers.go:53` `Env.Resource`（方法逻辑）<br>`render/render.go:102` 填充 `Env.Resources`（数据）<br>`static/static.go:156` `Write` 产出 map（产物） | **数据/逻辑/产物三处** |
| `helpers/fontawesome.rs::fa_helper` | — | FontAwesome 已删 |

### C. search.rs（445 行）→ internal/search + render/searchdocs.go

| Rust | Go 位置 |
|---|---|
| `create_files`（渲染器侧入口） | `render/render.go:445` `addSearchFiles`（调用方） |
| `index_chapter`（章节树 → 搜索文档） | `render/searchdocs.go:54` `indexChapter` |
| elasticlunr 索引构建（porter stemmer 等） | `search/index.go`、`pipeline.go`、`stemmer.go`、`search.go` |

### D. static_files.rs（320 行）→ internal/frontend（较完整；2026-08-18 由 internal/static 改名并去掉连字符）

| Rust | Go 位置 |
|---|---|
| `StaticFiles::new` | `static/static.go:59` `New` |
| `add_builtin` | `static.go:87` `AddBuiltin` |
| `hash_files` | `static.go:98` `Hash` |
| `write_files`（返回 resource helper closure） | `static.go:156` `Write`（返回 map → 进 Env.Resources） |

### E. handlebars 引擎 → tplgotpl（整包）

| Rust | Go |
|---|---|
| handlebars crate（**外部依赖**） | html/template（标准库）+ `tplgotpl` 薄封装（Registry/Env） |
| theme/*.hbs 模板（index/redirect/head/header/toc_js；toc_html 已删） | `tplgotpl/prod/` 1 个 .gohtml（index；2026-08-18 后 redirect/head/header 亦删） |

### F. 周边被拉动的

- `runner/build.go:29` `Build()` —— Rust 的 `impl Renderer for HtmlHandlebars` 本在
  hbs_renderer.rs 里，Go 直接由 runner 调 `render.Render`，没有 Renderer 接口对象
- `theme.go` —— Rust 的 Theme 结构有 **6 个模板字段**（index/head/redirect/
  header/toc_js/toc_html），Go 的 Theme 只有 assets/（CSS+JS）+ 1 个 gohtml
  模板字段（index；toc.js 改由 render.go 直接生成并 AddBuiltin）

## 3. 散落问题清单（聚合的动机）

1. **一个渲染器 = 7 个包**，且 `render` 包内部还混着编排/数据/搜索/JS 资产 4 件事
   （redirect 链路于 2026-08-18 整体移除）
2. **toc helper 拆 3 处**（算法/toc_render.go、方法/helpers.go、拼接/render.go；
   2026-08-15 后 in-page outline 那条拆点已删）
3. **resource helper 拆 3 处**（方法在 tplgotpl、数据在 render、产物在 static）
4. **search.rs 拆两半**：文档收集（render/searchdocs.go）和索引构建（search/）分离，
   `create_files` 的调用方还挂在 render 里
5. **toc.js 的 3 段 JS 常量**在 render.go 里（Rust 是 theme 模板）
6. **模板（tplgotpl/prod/）与渲染器（render/）分离**——Rust 里模板是 theme 字段、
   渲染器注册，一个 crate 内的事

## 4. ~~顺带发现的缺口：collect_redirects_for_path~~（已随 redirect 移除而作废，见顶部 2026-08-18 说明）

## 5. 聚合目标形态（建议，待确认）

聚合到 `internal/html`（= mdbook-html crate，与 html 管线合并）：

```
internal/html/
├── …现有 markdown 管线 10 文件（不动）
├── render.go      # hbs_renderer.rs：Render / renderChapter / render404 + buildTocJS + tocJSPrefix/Middle/AfterClass
├── data.go        # RenderData + makeData + chapterSummaries
├── search.go      # index_chapter / create_files（从 render/searchdocs.go 迁入）
├── static.go      # StaticFiles（从 internal/frontend 迁入；2026-08-18 改名）
├── helpers.go     # Env / Resource / TocHTML（从 tplgotpl 迁入）
├── toc_render.go  # renderTocSidebar(算法,从 toc.rs::RenderToc 移植)
└── templates/     # prod/*.gohtml index
```

未决问题：
1. **tplgotpl 去留**：作为"引擎包"保留（≈ Rust 的外部 handlebars crate），还是并入
   html（它其实是仓库内的 handlebars 替代品，并入后聚合才完整）？
2. **internal/search 去留**：并入 html（search.rs 本来就在 mdbook-html crate 内），
   还是保持独立包但被 html 调用？
3. **config 的 GitRepositoryIconClass**（make_data 拆出的小块）是否回迁？
4. **theme 的模板字段**是否回归（Rust Theme 有 6 个模板字段，Go 目前不归 theme）？
