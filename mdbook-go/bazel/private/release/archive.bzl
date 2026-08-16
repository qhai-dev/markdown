"""archive.bzl — devdoc release tarball repository_rule.

For each host platform (darwin-arm64, darwin-amd64, linux-amd64,
windows-amd64), expose a repository containing a `devdoc` binary that the
`devdoc` rule can consume.

v1 simplification: the repository copies a pre-built devdoc binary from
a path the user supplies. The user runs `make build` once in
mdbook-go/ to populate bin/devdoc, then either accepts the default
(symlink ./bazel/bin/devdoc → ../mdbook-go/bin/devdoc) or overrides
`devdoc_bin` in MODULE.bazel.

When the first GitHub release ships, swap the cp() call below for an
http_archive() pulling the platform-specific tarball. See DEVLOG.md §3.
"""

# The BUILD.bazel content generated inside each host repository. Exposes
# a single `devdoc` label that points at the staged binary via a
# native.genrule.
_BUILD_FILE_CONTENT = """\
genrule(
    name = "devdoc",
    srcs = ["bin/devdoc"],
    outs = ["devdoc_bin"],
    cmd = "cp $< $@ && chmod +x $@",
    executable = True,
    visibility = ["//visibility:public"],
)
"""

def _devdoc_repository_impl(repository_ctx):
    repo_root = str(repository_ctx.path(""))
    host_bin = repository_ctx.attr.devdoc_bin

    if not repository_ctx.path(host_bin).exists:
        msg = (
            "rules_devdoc: host-built devdoc binary not found at " + host_bin + ".\n" +
            "Run `make build` in mdbook-go/, then either symlink it to <workspace>/bazel/bin/devdoc\n" +
            "or override devdoc_bin in MODULE.bazel."
        )
        fail(msg)

    # Copy the host binary into this repository so each platform-specific
    # repo gets its own copy (and bazel can hash them independently).
    result = repository_ctx.execute(["mkdir", "-p", "{}/bin".format(repo_root)])
    if result.return_code != 0:
        fail("rules_devdoc: could not create bin dir: {}".format(result.stderr))
    result = repository_ctx.execute(["cp", host_bin, "{}/bin/devdoc".format(repo_root)])
    if result.return_code != 0:
        fail("rules_devdoc: could not stage devdoc binary: {}".format(result.stderr))
    result = repository_ctx.execute(["chmod", "+x", "{}/bin/devdoc".format(repo_root)])
    if result.return_code != 0:
        fail("rules_devdoc: could not chmod devdoc binary: {}".format(result.stderr))

    repository_ctx.file("BUILD.bazel", _BUILD_FILE_CONTENT)

    if hasattr(repository_ctx, "repo_metadata"):
        return repository_ctx.repo_metadata(reproducible = True)
    return None

devdoc_release_repository = repository_rule(
    implementation = _devdoc_repository_impl,
    attrs = {
        "devdoc_bin": attr.string(
            doc = "Path to the pre-built devdoc binary. Either absolute or relative to the main workspace (cwd when running bazel).",
            mandatory = True,
        ),
    },
    doc = "Materializes a devdoc binary for a single host platform by copying a pre-built binary.",
)