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

### Signed APT Repository

For Debian-like distributions such as deepin, UOS, Debian, Ubuntu, Kylin and Linux Mint:

```bash
sudo apt update
sudo apt install -y ca-certificates curl
sudo install -d -m 0755 /etc/apt/keyrings
curl -fsSL https://xynrin.github.io/spark-store-tui/spark-store-tui-archive-keyring.gpg | sudo tee /etc/apt/keyrings/spark-store-tui-archive-keyring.gpg >/dev/null
printf '%s\n' 'deb [signed-by=/etc/apt/keyrings/spark-store-tui-archive-keyring.gpg] https://xynrin.github.io/spark-store-tui stable main' | sudo tee /etc/apt/sources.list.d/spark-store-tui.list
sudo apt update
sudo apt install spark-store-tui
```

Signing key fingerprint:

```text
1AE6D4E7C4DB8C016F72F8C6A4D276F9CF8E57A9
```

### Local DEB

```bash
curl -LO https://github.com/Xynrin/spark-store-tui/releases/download/v0.7.2/spark-store-tui_0.7.2-1_all.deb
sudo apt install ./spark-store-tui_0.7.2-1_all.deb
```

### Non-APT Distributions

For Arch, Fedora, openSUSE and other non-Debian systems, install the runtime dependencies with your package manager, then install the script into `~/.local/bin`.

Native package manager recipes are included under `packaging/`. Arch users can
install from the AUR:

```bash
yay -S spark-store-tui
```

Package page: https://aur.archlinux.org/packages/spark-store-tui

After a Fedora COPR repository is published, Fedora users will be able to install
with:

```bash
sudo dnf copr enable xynrin/spark-store-tui
sudo dnf install spark-store-tui
```

Until the COPR package is published, Fedora users can use the self-hosted RPM
repository:

```bash
sudo tee /etc/yum.repos.d/spark-store-tui.repo >/dev/null <<'EOF'
[spark-store-tui]
name=Spark Store TUI
baseurl=https://xynrin.github.io/spark-store-tui/rpm
enabled=1
gpgcheck=0
repo_gpgcheck=0
EOF
sudo dnf install spark-store-tui
```

Fedora Atomic, Silverblue and Kinoite users can use the same repository, then
layer the package:

```bash
sudo rpm-ostree install spark-store-tui
systemctl reboot
```

If you prefer manual script installation, use the dependency commands below.

```bash
# Arch / Manjaro with yay
yay -S --needed bash curl jq fzf aria2 ca-certificates chafa

# Arch / Manjaro without yay
sudo pacman -S --needed bash curl jq fzf aria2 ca-certificates chafa

# Fedora Workstation / Server
sudo dnf install -y bash curl jq fzf aria2 ca-certificates chafa

# Fedora Atomic / Silverblue / Kinoite
sudo rpm-ostree install bash curl jq fzf aria2 ca-certificates chafa
systemctl reboot

# openSUSE
sudo zypper install -y bash curl jq fzf aria2 ca-certificates chafa
```

```bash
mkdir -p ~/.local/bin
curl -fsSL https://raw.githubusercontent.com/Xynrin/spark-store-tui/main/package-root/usr/bin/spark-store-tui -o ~/.local/bin/spark-store-tui
chmod +x ~/.local/bin/spark-store-tui
export PATH="$HOME/.local/bin:$PATH"
MODE=apm spark-store-tui
```

On non-Debian systems, `MODE=auto` defaults to APM Store metadata. Use `MODE=apm` explicitly if you want to avoid Spark Store `.deb` package content.

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

## Inspiration and Links

- Inspiration: [SHORiN-KiWATA/shorin-contrib pac](https://github.com/SHORiN-KiWATA/shorin-contrib/blob/main/pacman/pac)
- Friend link: [SHORiN-KiWATA](https://github.com/SHORiN-KiWATA)
- Spark Store official GitHub repository: [spark-store-project/spark-store](https://github.com/spark-store-project/spark-store)
- Spark Store official website: [spark-app.store](https://www.spark-app.store/)

## Project Information

- Author: Xynrin
- Maintainer: Xynrin <xynrin@163.com>
- Homepage: https://github.com/Xynrin/spark-store-tui
- Repository: https://github.com/Xynrin/spark-store-tui
