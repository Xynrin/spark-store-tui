# spark-store-tui

spark-store-tui is a terminal UI for browsing Spark Store and APM Store metadata. It is implemented with Bash, fzf, curl, jq and aria2c, and can show application lists, details, image previews and Metalink-based package downloads from the terminal.

License: GPL-3.0-only.
Author: Xynrin.
Maintainer: Xynrin <xynrin@163.com>.
Homepage: https://github.com/Xynrin/spark-store-tui.

## Features

- Bash + fzf terminal UI for store browsing.
- Spark Store and APM Store metadata browsing.
- Category, application list and application detail views.
- Image preview through chafa or viu when available.
- Metalink + aria2c package downloads.
- Temporary download directory cleanup by default.
- Debian-like systems default to Spark Store.
- Non-Debian systems default to APM Store only.
- amd64 and arm64 metadata paths are selected from `uname -m`.

## Install

### APT Repository

```bash
printf '%s\n' 'deb [trusted=yes] https://xynrin.github.io/spark-store-tui stable main' | sudo tee /etc/apt/sources.list.d/spark-store-tui.list
sudo apt update
sudo apt install spark-store-tui
```

### Local DEB

```bash
sudo apt install ./spark-store-tui_0.7.2-1_all.deb
```

## Run

```bash
spark-store-tui
```

## Test Download Without Installing

```bash
INSTALL_AFTER_DOWNLOAD=0 KEEP_DOWNLOADS=1 spark-store-tui
```

## Environment Variables

- `MODE`: `auto`, `spark`, `apm` or `choose`. The default is `auto`.
- `ARCH_PATH`: override the detected store path, for example `amd64-store`, `arm64-store`, `amd64-apm` or `arm64-apm`.
- `STRICT_SPEC`: when `1`, use documented Spark/APM metadata paths only. The default is `1`.
- `DIRECT_FALLBACK`: when `1`, try direct package URLs after Metalink download fails.
- `INSTALL_AFTER_DOWNLOAD`: `auto`, `1` or `0`. Use `0` to download without installing.
- `KEEP_DOWNLOADS`: when `1`, keep the download directory and print its path on exit.
- `DOWNLOAD_DIR`: custom download directory. Without this, downloads go to `/tmp/spark-store-tui.xxxxxx`.
- `IMAGE_PREVIEW`: when `1`, enable terminal image previews.
- `IMAGE_KIND`: `auto`, `icon` or `screen`. The default is `auto`.
- `IMAGE_WIDTH`: image preview width used by chafa or viu.
- `IMAGE_HEIGHT`: image preview height used by chafa.
- `ARIA2_READOUT`: when `0`, disable aria2c console readout. aria2c summaries are disabled by default.
- `ALLOW_SPARK_ON_NON_DEB`: when `1`, allow Spark Store mode on non-Debian systems for testing.

## Spark Store and APM Store

Spark Store mode uses Spark metadata and downloads from paths such as `https://d.spark-app.store/amd64-store/<category>/<pkgname>/app.json`. Downloaded packages are installed with `sudo ssinstall <local_file>` when installation is enabled and `ssinstall` is available.

APM Store mode uses APM metadata paths such as `https://d.spark-app.store/amd64-apm/<category>/<pkgname>/app.json`. Local package handling uses `sudo apm ssaudit <local_file>`, and the app also prints the online install hint `sudo apm install -y <pkgname>`.

## Automatic Store Selection

On Debian-like systems, including Deepin, UOS, Debian, Ubuntu, Kylin and Linux Mint, `MODE=auto` defaults to Spark Store.

On non-Debian systems, including Arch, Fedora, openSUSE, Manjaro and EndeavourOS, `MODE=auto` defaults to APM Store and Spark Store content is hidden. If `MODE=spark` is requested on a non-Debian system, the program switches to APM unless `ALLOW_SPARK_ON_NON_DEB=1` is set. `MODE=choose` follows the same rule: non-Debian systems only offer APM unless Spark is explicitly allowed.

## License

spark-store-tui is distributed under GPL-3.0-only. See `COPYING` for the full license text.

## Project Information

- Author: Xynrin
- Maintainer: Xynrin <xynrin@163.com>
- Homepage: https://github.com/Xynrin/spark-store-tui
- Repository: https://github.com/Xynrin/spark-store-tui
