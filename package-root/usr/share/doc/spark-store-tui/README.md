# Spark Store TUI

`sparkstore` is the primary command for Spark Store TUI 0.8.0.

The application reads Spark Store public metadata, downloads packages through
official Metalink mirrors, and can resume interrupted downloads after restart.
It supports native package-manager installation and removal after confirmation.

Optional dependencies:

- `chafa` for terminal image previews
- `sudo` when installing or removing packages as a non-root user

The package provides `sparkstore`, `SparkStore`, `SPARKSTORE`, and the legacy
`spark-store-tui` command alias.
