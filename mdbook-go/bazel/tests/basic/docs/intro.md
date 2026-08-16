# Welcome

This is the smoke-test fixture for rules_devdoc. The bazel action invokes
`devdoc build --dir <this directory>` and verifies the resulting HTML tree
contains the rendered intro page.