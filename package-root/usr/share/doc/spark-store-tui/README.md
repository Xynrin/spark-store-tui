# Spark Store TUI

`sparkstore` is the primary command for Spark Store TUI 0.8.0.

The application reads Spark Store public metadata. On Debian systems it uses
official aptss for download, resume, repository digest verification,
installation, selected-application updates and removal. Other supported hosts
use Amber APM after confirmation.

Optional dependencies:

- `chafa` for terminal image previews
- `sudo` when installing or removing packages as a non-root user

The package provides `sparkstore`, `SparkStore`, `SPARKSTORE`, and the legacy
`spark-store-tui` command alias.
