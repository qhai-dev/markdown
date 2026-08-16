"""Public API for rules_devdoc. Re-exports the two rules implemented under
//private/devdoc.bzl. See DEVLOG.md for the design rationale.
"""

load(
    "//private:devdoc.bzl",
    _devdoc = "devdoc",
    _devdoc_server = "devdoc_server",
)

devdoc = _devdoc
devdoc_server = _devdoc_server