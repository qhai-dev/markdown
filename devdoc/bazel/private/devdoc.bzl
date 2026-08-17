"""devdoc.bzl — the `devdoc` and `devdoc_server` rules.

These are the only public rules rules_devdoc exposes (v1). They correspond
to the upstream mdbook / mdbook_server rules from rules_rust_mdbook, with
two simplifications:

  1. No `process_wrapper` stage — devdoc reads absolute paths from bazel's
     sandbox directly, so the build action can call `devdoc build` with
     `--dir <book_root>` and trust the inputs are where they say they are.

  2. No toolchain query — the `devdoc` attribute is mandatory and points at
     the platform-specific host repository directly. bazel's toolchain
     registry is bypassed (see DEVLOG.md §4.3).
"""

DevdocInfo = provider(
    doc = "Information about a `devdoc` target.",
    fields = {
        "config": "File: The devdoc.yaml configuration file.",
        "srcs": "Depset[File]: Markdown sources passed as inputs.",
        "plugins": "Depset[File]: Preprocessor binaries injected via MDBOOK_PLUGIN_PATH (mirrors mdbook convention).",
    },
)

_PLUGINS_ENV = "MDBOOK_PLUGIN_PATH"

def _plugin_path(plugin, is_windows):
    sep = ";" if is_windows else ":"
    return sep.join(["${{pwd}}/{}".format(p.dirname) for p in plugin.to_list()])

def _devdoc_impl(ctx):
    book = ctx.file.book
    devdoc_bin = ctx.executable.devdoc

    output = ctx.actions.declare_directory(ctx.label.name)

    inputs = depset([book] + ctx.files.srcs)
    plugins = depset(ctx.files.plugins)

    is_windows = devdoc_bin.basename.endswith(".exe")
    plugin_path = _plugin_path(plugins, is_windows)

    # Mirror rules_rust_mdbook's process_wrapper pattern: run devdoc
    # with default build-dir (writes to <book_dir>/.devdoc/), then
    # copy the result into bazel's declared output. Passing --dest-dir
    # directly fails because devdoc writes to a sandbox-local path that
    # bazel discards once the action exits.
    cmd = "set -euo pipefail; '{devdoc}' build --dir '{bookdir}' && rm -rf '{output}' && mkdir -p '{output}' && cp -R '{bookdir}/.devdoc/.' '{output}/'".format(
        devdoc = devdoc_bin.path,
        bookdir = book.dirname,
        output = output.path,
    )

    ctx.actions.run_shell(
        command = cmd,
        inputs = inputs,
        tools = [devdoc_bin],
        outputs = [output],
        env = {_PLUGINS_ENV: plugin_path} if plugins else {},
        mnemonic = "DevdocBuild",
        progress_message = "devdoc build %{label}",
    )

    return [
        DefaultInfo(files = depset([output])),
        DevdocInfo(
            config = book,
            srcs = depset(ctx.files.srcs),
            plugins = depset(ctx.files.plugins),
        ),
    ]

devdoc = rule(
    implementation = _devdoc_impl,
    doc = "Build a devdoc book into a directory tree. The output is a single directory containing the rendered HTML site.",
    attrs = {
        "devdoc": attr.label(
            doc = "The devdoc binary to invoke. Must come from one of the rules_devdoc_toolchains_* repositories or a user-provided build target.",
            mandatory = True,
            allow_files = True,
            executable = True,
            cfg = "exec",
        ),
        "book": attr.label(
            doc = "The devdoc.yaml configuration file.",
            mandatory = True,
            allow_single_file = [".yaml"],
        ),
        "srcs": attr.label_list(
            doc = "All markdown sources and supporting files (theme, images, etc.) referenced by `devdoc.yaml`.",
            allow_files = True,
        ),
        "plugins": attr.label_list(
            doc = "Executables to inject into PATH for use as preprocessor commands. Wired through the MDBOOK_PLUGIN_PATH environment variable, matching the mdBook convention so existing preprocessors work unchanged.",
            allow_files = True,
            cfg = "exec",
        ),
    },
)

def _devdoc_server_impl(ctx):
    book_info = ctx.attr.book[DevdocInfo]
    devdoc_bin = ctx.executable.devdoc

    args = ctx.actions.args()
    args.add("--dir", book_info.config.dirname)
    args.add("serve")
    args.add("--hostname", ctx.attr.hostname)
    args.add("--port", ctx.attr.port)

    is_windows = devdoc_bin.basename.endswith(".exe")
    executable = ctx.actions.declare_file("{}{}".format(
        ctx.label.name,
        ".exe" if is_windows else "",
    ))
    ctx.actions.symlink(
        output = executable,
        target_file = devdoc_bin,
        is_executable = True,
    )

    return [
        DefaultInfo(
            executable = executable,
            files = depset([executable]),
            runfiles = ctx.runfiles(
                files = [book_info.config],
                transitive_files = depset(transitive = [book_info.srcs, book_info.plugins]),
            ),
        ),
        RunEnvironmentInfo(
            environment = {_PLUGINS_ENV: _plugin_path(book_info.plugins, is_windows)} if book_info.plugins else {},
        ),
    ]

devdoc_server = rule(
    implementation = _devdoc_server_impl,
    doc = "Spawn a devdoc server for a given `devdoc` target. Wraps `devdoc serve` so the bazel-built sources are served without needing the user to invoke devdoc manually.",
    attrs = {
        "book": attr.label(
            doc = "The `devdoc` target to serve.",
            providers = [DevdocInfo],
            mandatory = True,
        ),
        "devdoc": attr.label(
            doc = "The devdoc binary to invoke. Must match the platform the server will run on.",
            mandatory = True,
            allow_files = True,
            executable = True,
            cfg = "exec",
        ),
        "hostname": attr.string(
            doc = "Default hostname. Overridable on the command line via --hostname.",
            default = "localhost",
        ),
        "port": attr.string(
            doc = "Default port. Overridable on the command line via --port.",
            default = "3000",
        ),
    },
    executable = True,
)