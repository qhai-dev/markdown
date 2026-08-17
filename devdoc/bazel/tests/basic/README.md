# rules_devdoc basic smoke test

Run with:

    cd mdbook-go/bazel
    bazel test //tests/basic:build_test

This fixture is independent of `mdbook-go/examples/basic` (which is the
devdoc Go test fixture) so a regression in either side can be investigated
without disturbing the other.