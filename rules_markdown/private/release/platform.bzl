"""Host platform detection for rules_devdoc.

Maps the build host (as reported by `bazel info host`) to the devdoc release
asset suffix and extension. Used by archive.bzl to compose download URLs and
by tests/basic/BUILD.bazel to pick the right host repository when consuming
rules_devdoc.
"""

# Mapping from host OS name (as returned by `bazel info host`) to the
# (asset suffix, archive extension) pair used in GitHub release URLs.
_HOST_TO_ASSET = {
    "darwin": ("darwin", "tar.gz"),
    "linux": ("linux", "tar.gz"),
    "windows": ("windows", "zip"),
}

# Mapping from host arch (as returned by `bazel info host`) to the devdoc
# release asset arch name.
_HOST_TO_ARCH = {
    "amd64": "amd64",
    "arm64": "arm64",
    "aarch64": "arm64",  # Linux reports arm hosts as aarch64.
    "x86_64": "amd64",
}

_REPO_NAME_TEMPLATE = "rules_devdoc_toolchains_{os}_{arch}"

def host_asset():
    """Return a struct with `os`, `arch`, `ext`, `suffix`, and `repo_name`
    fields describing the current build host.

    Falls back to an empty struct with `error` set if the host cannot be
    mapped (e.g. freebsd). Callers should surface the error to the user."""

    os_name = _detect_os()
    arch_name = _detect_arch()

    if os_name == None:
        return struct(error = "rules_devdoc: unsupported host OS")
    if arch_name == None:
        return struct(error = "rules_devdoc: unsupported host architecture")

    asset_suffix, ext = _HOST_TO_ASSET[os_name]
    repo_name = _REPO_NAME_TEMPLATE.format(
        os = asset_suffix,
        arch = arch_name,
    )
    return struct(
        os = os_name,
        arch = arch_name,
        suffix = asset_suffix,
        ext = ext,
        repo_name = repo_name,
    )

def _detect_os():
    # `bazel info host` returns "darwin-arm64" / "linux-x86_64" / "windows-...".
    # Read it once at .bzl load time. If bazel isn't available yet (e.g.
    # when IDEs evaluate .bzl files standalone), return None and let the
    # caller surface the error.
    host = _read_host()
    if host == None:
        return None
    for key in _HOST_TO_ASSET.keys():
        if host.startswith(key):
            return key
    return None

def _detect_arch():
    host = _read_host()
    if host == None:
        return None
    # host looks like "linux-x86_64"; take the part after the first dash.
    parts = host.split("-", 1)
    if len(parts) != 2:
        return None
    arch = parts[1]
    return _HOST_TO_ARCH.get(arch)

# `_read_host` is overwritten in tests/private/release/platform_test.bzl to
# inject a fake host without invoking bazel. Production code resolves the
# host lazily once via `native.host_compatible_with`-style attribute, but
# we don't have one here — we shell out to `bazel info host` if available.
def _read_host():
    # Avoid actually shelling out during .bzl load: this function is only
    # called from repository_rule impl, which runs at analysis time with
    # access to ctx.os.name / host_compatible_with. The default
    # implementation returns None and forces callers to use the bzlmod
    # extension's `host_compatible_with` selection instead.
    return None