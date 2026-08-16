# rules_devdoc — 设计文档

> 状态：**设计阶段**。本文档定义 `rules_devdoc` 的形状、关键决策与拆分路径。
> 仓库内目前仅有此文件；任何 `BUILD.bazel` / `MODULE.bazel` / `*.bzl` 都在 review 通过本文档后补建。

---

## 0. 目标

为 [devdoc](https://github.com/qhai-dev/mdBook)（本仓库 `mdbook-go/` 子项目的 CLI 名称）提供一套 Bazel 规则，使用户能在自己的 workspace 里以增量、缓存友好的方式构建 devdoc 文档站点，对标 [rules_rust_mdbook](https://github.com/bazelbuild/rules_rust/tree/main/extensions/mdbook) 的能力与 API。

---

## 1. 与参照物的对应关系

| 维度 | rules_rust_mdbook（参照） | rules_devdoc（本设计） |
|---|---|---|
| 模块名 | `rules_rust_mdbook` | `rules_devdoc` |
| 公共规则 | `mdbook` / `mdbook_server` / `mdbook_toolchain` | `devdoc` / `devdoc_server`（`devdoc_toolchain` rule **v1 不实现**，见 §4.3） |
| 工具链来源 | `rust_ext` 拉 `rmdbi__mdbook-0.4.44` 预编译 crate | GitHub release tarball + `repository_rule` 拉取 |
| 平台选择 | rust_ext 内部按 host 选 | `select({...})` + `host_compatible_with` 显式声明 |
| 配置文件 | `book.toml` (`allow_single_file = ["book.toml"]`) | `devdoc.yaml` (`allow_single_file = [".yaml"]`) |
| 配置 root | `<book dir>/book.toml` 同级找 `src/` | `<book dir>/devdoc.yaml` 同级找 `docs/`（devdoc 默认） |
| Process wrapper | `private/process_wrapper.rs`（Rust） | **省略**——devdoc 直接读沙箱绝对路径，无需 stage |
| Server | `private/server.rs`（Rust） | `private/server.go`（Go），复用 devdoc 的 polling watch |
| Plugins | `MDBOOK_PLUGIN_PATH` 环境变量注入 | 同左（devdoc 协议兼容 mdbook preprocessor） |
| 私有 bzl 库 | `private/bzl_lib` | 同左，名字 `private/bzl_lib` |
| Provider | `MdBookInfo(config, srcs, plugins)` | `DevdocInfo(config, srcs, plugins)` |

### 1.1 关键简化

**不实现 process_wrapper stage**：devdoc 用 `os.ReadFile` 读绝对路径，不假设工作目录有源文件副本。意味着：

- bazel action 把所有 inputs 放进沙箱后，`devdoc build --dir <book_root>` 即可，`book_root` 是 `book.dirname`
- **代价**：devdoc 不能依赖「相对当前工作目录」的任何解析逻辑。我们已经验证（见 `mdbook-go/internal/runner/loader.go::Load`）`Load(root)` 完全基于 `root` 解析所有路径，符合预期。

**不实现 server.rs 完整 ibazel 协议**（v1）：v1 server 只做 stage + `devdoc serve`，ibazel 信号作为 follow-up。这与参照物对称——参照物的 `server.rs` 也是独立组件，不影响 `mdbook` 规则本身。

---

## 2. 文件结构（最终形态）

```
mdbook-go/bazel/
├── README.md                  # 给 user 看：如何在他们的 workspace 用
├── BUILD.bazel                # 顶层公共目标
├── MODULE.bazel               # bzlmod 模块声明
├── defs.bzl                   # 公共门面
└── private/
    ├── BUILD.bazel            # 内部 bzl_library
    ├── devdoc.bzl             # devdoc / devdoc_server 实现
    └── release/
        ├── archive.bzl        # repository_rule: 拉 GitHub release tarball
        └── platform.bzl       # host 平台 select + URL 模板
```

**10 个文件**，与参照物 8 个文件（不含 `process_wrapper.rs` / `server.rs` / `internal_extensions.bzl` / `.bazelrc` / `.bazelignore` / `.gitignore`）的差异：

- `private/release/archive.bzl`、`private/release/platform.bzl` 两个新文件（参照物用 `rust_ext`，devdoc 没有等价物）
- `private/server.go`（参照物的 `server.rs`）—— **v1 范围外**，见 §5
- `private/process_wrapper.rs` / `process_wrapper` 目标 **省略**

---

## 3. 模块元数据（MODULE.bazel 草稿）

```python
"""rules_devdoc — Bazel rules for devdoc (https://github.com/qhai-dev/mdBook)."""

module(
    name = "rules_devdoc",
    version = "0.1.0",
)

bazel_dep(name = "platforms", version = "1.1.0")
bazel_dep(name = "bazel_skylib", version = "1.8.2")

# devdoc 工具链的 release tarball 拉取规则在 private/release/archive.bzl 定义。
# 注意：v1 不调用 register_toolchains()——devdoc rule 直接通过 mandatory devdoc attr
# 引用具体 host 平台的二进制仓库，不走 bazel 工具链注册表（见 §4.3）。
archive_ext = use_extension("@rules_devdoc//private/release:extensions.bzl", "archive_ext")
use_repo(
    archive_ext,
    "rules_devdoc_toolchains_linux_amd64",
    "rules_devdoc_toolchains_darwin_amd64",
    "rules_devdoc_toolchains_darwin_arm64",
)
```

### 3.1 release tarball 契约

devdoc CI 在每次 `git tag vX.Y.Z` 后向 GitHub release 上传：

```
devdoc-vX.Y.Z-linux-amd64.tar.gz
devdoc-vX.Y.Z-darwin-amd64.tar.gz
devdoc-vX.Y.Z-darwin-arm64.tar.gz
devdoc-vX.Y.Z-windows-amd64.zip
```

每个 tarball 内固定结构：

```
devdoc/
└── bin/devdoc          # 可执行文件
```

`archive.bzl::devdoc_release_repository` 解析 URL 模板：

```
https://github.com/qhai-dev/mdBook/releases/download/v{version}/devdoc-v{version}-{platform}.{ext}
```

`{platform}` / `{ext}` 由 `platform.bzl` 的 `_host_platform()` 决定。

### 3.2 fallback：release 不存在时

首个 release 之前，`archive_ext` 提供 fallback 走 `go install`：

```python
fallback_version = "v0.0.0-dev"  # 任何不存在的 tag，触发 fallback 路径
```

fallback 实现：在 `_devdoc_release_repository_impl` 里 `exec(["go", "install", "github.com/qhai-dev/mdBook/cmd/devdoc@latest"])`，把生成的 `$GOBIN/devdoc` 包成相同形状的 repository。**发布第一个 GitHub release 后移除此分支**。

---

## 4. 规则 API（defs.bzl 草稿）

### 4.1 `devdoc`

```python
devdoc(
    name = "book",
    devdoc = "@rules_devdoc_toolchains_darwin_arm64//:devdoc",  # 工具二进制来源（mandatory）
    book = "devdoc.yaml",           # 配置入口（devdoc.yaml）
    srcs = glob(["docs/**/*.md"]),  # 所有源文件
    plugins = [":my_preprocessor"], # 可选：bazel 目标作为 preprocessor 二进制
)
```

**`devdoc` attr 是 mandatory**：用户必须显式指定 devdoc 二进制来源（来自哪个 host 平台的 bzlmod 仓库）。理由：
- 显式优于隐式——用户清楚知道 bazel action 跑的是哪个二进制
- 跨平台 CI 测试方便（同一 BUILD 在 darwin 和 linux runner 上分别指向不同 host 仓库）
- 失去的能力：「bazel 自动按 host 平台挑选」——但 bzlmod extension 在 MODULE.bazel 里 `select` 一样能达成

规则实现要点（对照 `rules_rust_mdbook/private/mdbook.bzl::_mdbook_impl`）：

| 步骤 | 参照物 | devdoc |
|---|---|---|
| 1. 拉工具链 | `toolchains["@rules_rust_mdbook//:toolchain_type"]` | **不需要**——`devdoc` rule 直接读 `ctx.file.devdoc`（来自 `devdoc` attr） |
| 2. 计算 plugin path | `MDBOOK_PLUGIN_PATH` 环境变量 | 同左，devdoc 协议兼容 |
| 3. inputs map | `_map_inputs(file)` + `--argfile` | **省略**（见 §1.1） |
| 4. 调工具 | `process_wrapper` 间接调 `mdbook build` | 直接调 `devdoc build --dir <book_root>` |
| 5. 输出 | `ctx.actions.declare_directory(name)` | 同左 |

### 4.2 `devdoc_server`

```python
devdoc_server(
    name = "book_server",
    book = ":book",
    hostname = "localhost",   # 默认
    port = "3000",            # 默认
    tags = ["ibazel_notify_changes"],  # 可选
)
```

v1 行为：调 `devdoc serve --dir <book_root>`，server 不参与 stage（devdoc polling 监听绝对路径）。ibazel 热重载信号处理留 v2。

### 4.3 `devdoc_toolchain`（v1 不实现）

**v1 跳过这个 rule**。理由：
- devdoc 工具链不需要 user 自定义（不像 Rust 工具链那样 user 经常 fork）
- `devdoc` rule 已经直接接受 `devdoc` attr 做工具来源，多一层 `devdoc_toolchain` rule 是纯间接
- 失去的能力：user 在自己 workspace 里 `register_toolchains` 覆盖默认工具链——这个能力 v1 用不到

**如果 v2 需要**（比如 devdoc 引入第三方 preprocessor 编译成独立二进制、要求用户能在自己 workspace 提供修补版 devdoc），再补 `_devdoc_toolchain_impl` + `current_devdoc_binary` + `devdoc_toolchain` rule。`toolchain_type` 也届时一并注册——v1 完全不需要。

---

## 5. v1 / v2 / v3 范围切片

### v1（本次实现，必须完整）

- `MODULE.bazel` + 两个公共规则（`devdoc` / `devdoc_server`），不实现 `devdoc_toolchain` rule
- release tarball 拉取 + fallback
- 一个测试夹具：`tests/basic/` 用 `examples/basic/` 作 fixture，跑 `bazel build //bazel/tests/basic:book`
- 文档：`README.md` + 本 `DEVLOG.md`

### v2（follow-up PR）

- `devdoc_server` 完整 ibazel 协议（stdin 信号 → 重启子进程）
- `private/server.go`：stage + `devdoc serve` + signal handling
- 更多测试夹具：`plugins/`、`external_srcs/`、`generated_srcs/`，对照参照物 `test/`

### v3（v2 之后）

- `rules_mdbook` 占位升级为完整实现（如果用户需要在自己 workspace 构建 Rust mdBook 文档）
- 与 bcr 发布流水线对接

---

## 6. 测试策略

### 6.1 v1 夹具

`bazel/tests/basic/BUILD.bazel`：

```python
load("@rules_devdoc//:defs.bzl", "devdoc")
load("@bazel_skylib//rules:build_test.bzl", "build_test")

devdoc(
    name = "book",
    book = "//examples/basic:devdoc.yaml",  # 借用 devdoc 自己的 examples
    srcs = glob(["docs/**/*.md"]),
)

build_test(
    name = "build_test",
    targets = [":book"],
)
```

**问题**：`examples/basic` 是 devdoc 的 Go 测试夹具，被 Go 测试消费。在 bazel 夹具里复用意味着引入跨语言依赖。**v1 的替代方案**：新建 `bazel/tests/basic/` 自己的 docs/ 目录，含一个最小 `intro.md` + `devdoc.yaml`，确保 devdoc 的 Go 夹具和 bazel 夹具**不共享文件**。

### 6.2 验证步骤

```bash
cd mdbook-go/
bazel build //bazel/tests/basic:book
# 期望：产出 bazel/tests/basic/bazel-out/.../book/ 目录
# 验证：bazel-bin/bazel/tests/basic/book/index.html 存在且非空
```

---

## 7. 拆分路径：monorepo → 独立模块

当前物理位置 `mdbook-go/bazel/` **不是**独立 bazel 模块（bazel 按仓库根的 `MODULE.bazel` 识别模块）。要从 monorepo 拆出独立模块：

1. **新建仓库** `qhai-dev/rules_devdoc`（GitHub repo）
2. **复制**：把 `mdbook-go/bazel/` 下所有文件复制到新仓库根
3. **调整路径**：所有引用 `//private/...` 的地方改为相对新仓库根的 `//private/...`，路径不变
4. **调整 module 声明**：`MODULE.bazel` 移到新仓库根（原样即可，因为模块名是 `rules_devdoc` 不依赖父路径）
5. **首次 bcr 发布**：`rules_devdoc/v0.1.0` 提交到 bcr candidate pool

**关键不变量**：上述拆分**不需要改动任何 `.bzl` / `.bazel` 文件的内容**，只是物理位置变化。所有 `//private/release:extensions.bzl` 等路径在新仓库根下仍然正确。

### 7.1 拆分时容易踩的坑

- **bzl_library visibility**：`private/bzl_lib` 默认 `visibility = ["//:__pkg__"]`，要求 bzl_library 与消费者在同仓库。这是我们的设计，拆分后不变。
- **平台 select 的 cpu 值**：本设计沿用 `platforms//host` 的标准约束。`rules_rust_mdbook` 用了 `@rules_rust//util:platforms`，devdoc 没这个依赖，所以 `platforms` 直接从 bcr 拉，**没有**自定义 cpu 值。
- **`@rules_devdoc` 自己的仓库名**：用户 `bazel_dep` 它时用的是模块名（`rules_devdoc`）而不是仓库名（`rules_devdoc`），因为我们 `module(name = "rules_devdoc", ...)`。如果 GitHub repo 叫别的名（比如 `rules_devdoc_archive`），需要在 README 里说明这是 bzlmod 模块名 vs GitHub repo 名。

---

## 8. 风险与未决项

### 8.1 风险

| 风险 | 缓解 |
|---|---|
| devdoc 0.5.x 后改 `Load` 接受 `devdoc.yaml` 之外的路径 | 锁定 devdoc 版本到首次发布时的 tag |
| macOS arm64 release 缺失导致本地不可用 | fallback 走 `go install`（§3.2） |
| bzlmod 与 WORKSPACE 兼容 | v1 仅支持 bzlmod，WORKSPACE 用户在 README 显式声明 |
| devdoc `serve` 模式对 ibazel stdin 信号无响应 | v1 不实现，v2 再做 |

### 8.2 未决项

1. **devdoc 自身是否需要 release 脚本**？（暂定：是，建议 `.github/workflows/release.yml` 在 `git tag v*` 触发，build matrix 出 4 个平台 tarball）
2. **`rules_mdbook` 占位文件的保留期**？（暂定：v3 之前占位不撤）
3. **server.go 复用 devdoc 的 polling 还是用 inotify**？（v2 决策，取决于 devdoc 是否暴露 `--no-poll` flag）

---

## 9. 行动项（按依赖排序）

- [ ] **A1** 创建 `mdbook-go/bazel/` 目录与本 `DEVLOG.md`（本次完成）
- [ ] **A2** user review 本文档 → 确认 v1 范围、fallback 策略、测试策略
- [ ] **A3** 起草 `MODULE.bazel`、`defs.bzl`、`private/release/archive.bzl`、`private/release/platform.bzl`
- [ ] **A4** 起草 `private/devdoc.bzl`（仅 `devdoc` + `devdoc_server` 实现，不写 `devdoc_toolchain`）、`private/release/archive.bzl`、`private/release/platform.bzl`、`private/BUILD.bazel`、`BUILD.bazel`
- [ ] **A5** `bazel/tests/basic/` 夹具 + `bazel build` 验证
- [ ] **A6** devdoc release 脚本（`hack/release.sh` 或 GitHub Actions）
- [ ] **A7** 拆 repo + 首次 bcr candidate 提交

A2 完成后 A3–A5 是同一 PR；A6 与 A7 是发布日的工作。