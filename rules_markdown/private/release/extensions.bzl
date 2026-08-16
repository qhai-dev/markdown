"""extensions.bzl — bzlmod extension that exposes per-host devdoc repositories.

Usage from MODULE.bazel:
    archive_ext = use_extension("@rules_devdoc//private/release:extensions.bzl", "archive_ext")
    archive_ext.devdoc_bin(path = "/path/to/bin/devdoc")
    use_repo(archive_ext, "rules_devdoc_toolchains_darwin_arm64", ...)

If `archive_ext.devdoc_bin(path = ...)` is omitted, the default path
`<bazel execroot>/bin/devdoc` is used — i.e. the user must run
`make build` in mdbook-go/ from the same directory they invoke bazel.
"""

load("//private/release:archive.bzl", "devdoc_release_repository")

# Host matrix. Each entry produces one bzlmod repository.
_HOSTS = [
    # (repo_name, asset_suffix)
    ("rules_devdoc_toolchains_linux_amd64", "linux-amd64"),
    ("rules_devdoc_toolchains_darwin_amd64", "darwin-amd64"),
    ("rules_devdoc_toolchains_darwin_arm64", "darwin-arm64"),
]

# Default devdoc binary path. Resolved relative to the bazel invocation
# directory at fetch time. The user runs `make build` in mdbook-go/
# from the same directory before invoking bazel.
_DEFAULT_DEVDOC_BIN = "bin/devdoc"

def _archive_ext_impl(module_ctx):
    devdoc_bin = _DEFAULT_DEVDOC_BIN
    for mod in module_ctx.modules:
        for tag in mod.tags.devdoc_bin:
            if tag.path:
                devdoc_bin = tag.path

    for repo_name, asset in _HOSTS:
        devdoc_release_repository(
            name = repo_name,
            devdoc_bin = devdoc_bin,
        )
    return module_ctx.extension_metadata(
        root_module_direct_deps = [repo_name for repo_name, _ in _HOSTS],
        root_module_direct_dev_deps = [],
    )

archive_ext = module_extension(
    implementation = _archive_ext_impl,
    doc = "Exposes one devdoc binary repository per host platform.",
    tag_classes = {
        "devdoc_bin": tag_class(attrs = {
            "path": attr.string(
                doc = "Path to the pre-built devdoc binary. Default: bin/devdoc (resolved relative to the bazel invocation cwd).",
                default = "",
            ),
        }),
    },
)