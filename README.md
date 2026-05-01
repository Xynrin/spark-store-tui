<p align="center">
  <img src="icon.png" width="128" alt="Spark Store TUI icon">
</p>

# Spark Store TUI / 星火终端助手

![License: GPL-3.0-only](https://img.shields.io/badge/License-GPL--3.0--only-blue.svg)
![AUR version](https://img.shields.io/aur/version/spark-store-tui)
![GitHub release](https://img.shields.io/github/v/release/Xynrin/spark-store-tui)
![Shell](https://img.shields.io/badge/Shell-Bash-4EAA25)
![TUI](https://img.shields.io/badge/TUI-fzf-orange)

Spark Store TUI, 中文名「星火终端助手」，is a Bash + fzf terminal UI for browsing Spark Store and APM Store metadata. It can show category lists, app details, terminal image previews, and Metalink-based package downloads from the terminal.

This is a third-party terminal tool, not the official Spark Store client.

| Item | Value |
|---|---|
| Chinese name | 星火终端助手 |
| English name | Spark Store TUI |
| Package name | `spark-store-tui` |
| Current package version | `0.7.2-1` |
| Command | `spark-store-tui` |
| License | GPL-3.0-only |
| Release channels | deb / rpm / AUR / source |
| GitHub | https://github.com/Xynrin/spark-store-tui |
| Gitee mirror | https://gitee.com/spark-store-project/spark-store-tui |

## Features

- Bash + fzf terminal UI for store browsing.
- Spark Store and APM Store metadata browsing.
- Category, application list and application detail views.
- Image preview through chafa or viu when available.
- Metalink + aria2c package downloads.
- Debian-like systems default to Spark Store.
- Non-Debian systems default to APM Store only.
- amd64 and arm64 metadata paths are selected from `uname -m`.
- Downloads go to `/tmp/spark-store-tui.xxxxxx` by default and are cleaned up on exit.
- `KEEP_DOWNLOADS=1` keeps the download directory and prints its path.

## Install

### Debian / deepin / UOS / Ubuntu / Linux Mint

Use the signed APT repository hosted by the Gitee organization mirror:

```bash
sudo apt update
sudo apt install -y ca-certificates curl
sudo install -d -m 0755 /etc/apt/keyrings
curl -fsSL https://gitee.com/spark-store-project/spark-store-tui/raw/master/apt/spark-store-tui-archive-keyring.gpg | sudo tee /etc/apt/keyrings/spark-store-tui-archive-keyring.gpg >/dev/null
printf '%s\n' 'deb [signed-by=/etc/apt/keyrings/spark-store-tui-archive-keyring.gpg] https://gitee.com/spark-store-project/spark-store-tui/raw/master/apt stable main' | sudo tee /etc/apt/sources.list.d/spark-store-tui.list
sudo apt update
sudo apt install spark-store-tui
```

Signing key fingerprint:

```text
1AE6D4E7C4DB8C016F72F8C6A4D276F9CF8E57A9
```

Install the `.deb` package directly:

```bash
curl -LO https://github.com/Xynrin/spark-store-tui/releases/download/v0.7.2/spark-store-tui_0.7.2-1_all.deb
sudo apt install ./spark-store-tui_0.7.2-1_all.deb
```

Domestic mirror:

```bash
curl -LO https://gitee.com/spark-store-project/spark-store-tui/raw/master/apt/pool/main/s/spark-store-tui/spark-store-tui_0.7.2-1_all.deb
sudo apt install ./spark-store-tui_0.7.2-1_all.deb
```

### Arch / Manjaro / EndeavourOS

The package is available on AUR. `yay` searches Arch repositories and AUR, not ArchWiki.

```bash
yay -S spark-store-tui
```

Package page: https://aur.archlinux.org/packages/spark-store-tui

### Fedora / RPM

Use the self-hosted RPM repository:

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

Domestic Gitee mirror / 国内 Gitee 镜像：

```bash
sudo tee /etc/yum.repos.d/spark-store-tui.repo >/dev/null <<'EOF'
[spark-store-tui]
name=Spark Store TUI
baseurl=https://gitee.com/spark-store-project/spark-store-tui/raw/master/rpm
enabled=1
gpgcheck=0
repo_gpgcheck=0
EOF
sudo dnf install spark-store-tui
```

Fedora Atomic, Silverblue and Kinoite users can use the same repository, then layer the package:

```bash
sudo rpm-ostree install spark-store-tui
systemctl reboot
```

### Manual Script Install

For openSUSE and other non-APT distributions, install dependencies with your package manager first:

```bash
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

Then install the script:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://raw.githubusercontent.com/Xynrin/spark-store-tui/main/package-root/usr/bin/spark-store-tui -o ~/.local/bin/spark-store-tui
chmod +x ~/.local/bin/spark-store-tui
export PATH="$HOME/.local/bin:$PATH"
MODE=apm spark-store-tui
```

Domestic mirror:

```bash
mkdir -p ~/.local/bin
curl -fsSL https://gitee.com/spark-store-project/spark-store-tui/raw/master/package-root/usr/bin/spark-store-tui -o ~/.local/bin/spark-store-tui
chmod +x ~/.local/bin/spark-store-tui
export PATH="$HOME/.local/bin:$PATH"
MODE=apm spark-store-tui
```

On non-Debian systems, `MODE=auto` defaults to APM Store metadata. Use `MODE=apm` explicitly if you want to avoid Spark Store `.deb` package content.

## Run

```bash
spark-store-tui
```

Download without installing and keep the download directory:

```bash
INSTALL_AFTER_DOWNLOAD=0 KEEP_DOWNLOADS=1 spark-store-tui
```

## Environment Variables

| Variable | Description |
|---|---|
| `MODE` | `auto`, `spark`, `apm` or `choose`; default: `auto` |
| `ARCH_PATH` | Override the detected store path, for example `amd64-store`, `arm64-store`, `amd64-apm` or `arm64-apm` |
| `STRICT_SPEC` | Use documented Spark/APM metadata paths only when set to `1`; default: `1` |
| `DIRECT_FALLBACK` | Try direct package URLs after Metalink download fails when set to `1` |
| `INSTALL_AFTER_DOWNLOAD` | `auto`, `1` or `0`; use `0` to download without installing |
| `KEEP_DOWNLOADS` | Keep the download directory and print its path on exit when set to `1` |
| `DOWNLOAD_DIR` | Custom download directory; otherwise `/tmp/spark-store-tui.xxxxxx` is used |
| `IMAGE_PREVIEW` | Enable terminal image previews when set to `1` |
| `IMAGE_KIND` | `auto`, `icon` or `screen`; default: `auto` |
| `IMAGE_WIDTH` | Image preview width used by chafa or viu |
| `IMAGE_HEIGHT` | Image preview height used by chafa |
| `ARIA2_READOUT` | Disable aria2c console readout when set to `0`; aria2c summaries are disabled by default |
| `ALLOW_SPARK_ON_NON_DEB` | Allow Spark Store mode on non-Debian systems for testing when set to `1` |

## Spark Store and APM Store

Spark Store mode uses Spark metadata and downloads from paths such as:

```text
https://d.spark-app.store/amd64-store/<category>/<pkgname>/app.json
```

Downloaded packages are installed with `sudo ssinstall <local_file>` when installation is enabled and `ssinstall` is available.

APM Store mode uses APM metadata paths such as:

```text
https://d.spark-app.store/amd64-apm/<category>/<pkgname>/app.json
```

Local package handling uses `sudo apm ssaudit <local_file>`, and the app also prints the online install hint:

```bash
sudo apm install -y <pkgname>
```

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
- GitHub: https://github.com/Xynrin/spark-store-tui
- Gitee mirror: https://gitee.com/spark-store-project/spark-store-tui
