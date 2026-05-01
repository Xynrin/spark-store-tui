<p align="center">
  <img src="icon.png" width="128" alt="星火终端助手图标">
</p>

# 星火终端助手

星火终端助手（Spark Store TUI）是一个基于 Bash、fzf、curl、jq、aria2c 的终端 TUI 应用商店浏览工具，可以在终端里浏览 Spark Store 与 APM Store 的应用元数据、查看应用详情，并通过 Metalink 与 aria2c 下载软件包。

这是第三方终端工具，不是星火应用商店官方客户端。

| 项目 | 状态 |
|---|---|
| 中文名 | 星火终端助手 |
| 英文名 | Spark Store TUI |
| 包名 | spark-store-tui |
| 当前版本 | 0.7.2-1 |
| 许可证 | GPL-3.0-only |
| 运行方式 | 终端命令 |
| 支持平台 | Debian-like / APM-compatible |
| 发布格式 | deb / rpm / AUR / source |
| Gitee 仓库 | https://gitee.com/spark-store-project/spark-store-tui |
| GitHub 备用源码 | https://github.com/Xynrin/spark-store-tui |

## 功能特性

- 使用 Bash + fzf 实现终端交互界面。
- 支持浏览 Spark Store 与 APM Store 元数据。
- 支持分类列表、应用列表、应用详情页面。
- 安装 chafa 或 viu 后可在终端中预览图片。
- 使用 Metalink + aria2c 下载软件包。
- Debian-like 系统默认显示 Spark Store 内容。
- 非 Debian 系统默认只显示 APM Store 内容，避免展示不兼容的 `.deb` 包。
- 根据 `uname -m` 自动选择 amd64 / arm64 等元数据路径。
- 默认下载到 `/tmp/spark-store-tui.xxxxxx`，退出时自动清理。
- 设置 `KEEP_DOWNLOADS=1` 可以保留下载目录。

## Debian / deepin / UOS / Ubuntu / Linux Mint

### 使用 APT 源安装

本 Gitee 组织仓库提供自建 APT 源，适合国内网络环境。APT 源已配置仓库签名。

```bash
sudo apt update
sudo apt install -y ca-certificates curl
sudo install -d -m 0755 /etc/apt/keyrings
curl -fsSL https://gitee.com/spark-store-project/spark-store-tui/raw/master/apt/spark-store-tui-archive-keyring.gpg | sudo tee /etc/apt/keyrings/spark-store-tui-archive-keyring.gpg >/dev/null
printf '%s\n' 'deb [signed-by=/etc/apt/keyrings/spark-store-tui-archive-keyring.gpg] https://gitee.com/spark-store-project/spark-store-tui/raw/master/apt stable main' | sudo tee /etc/apt/sources.list.d/spark-store-tui.list
sudo apt update
sudo apt install spark-store-tui
```

APT 签名公钥指纹：

```text
1AE6D4E7C4DB8C016F72F8C6A4D276F9CF8E57A9
```

### 直接安装 deb 包

```bash
curl -LO https://gitee.com/spark-store-project/spark-store-tui/raw/master/apt/pool/main/s/spark-store-tui/spark-store-tui_0.7.2-1_all.deb
sudo apt install ./spark-store-tui_0.7.2-1_all.deb
```

## Arch / Manjaro / EndeavourOS

AUR 包已经发布，可以用 yay 直接安装。yay 搜索的是 Arch 官方仓库与 AUR，不是 ArchWiki。

```bash
yay -S spark-store-tui
```

AUR 页面：https://aur.archlinux.org/packages/spark-store-tui

## Fedora / openSUSE / 其他非 APT 发行版

### Fedora DNF 自建 RPM 源

当前仓库提供自建 RPM 源，可用 `dnf install spark-store-tui` 安装依赖和软件包。这个 RPM 源目前未启用 GPG 签名校验。

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

Fedora Atomic、Silverblue、Kinoite 可以使用同一个 RPM 源，然后通过 rpm-ostree 分层安装：

```bash
sudo rpm-ostree install spark-store-tui
systemctl reboot
```

### 手动脚本安装

如果你的发行版暂时没有可用包，可以先安装运行依赖，再把脚本放到 `~/.local/bin`。

```bash
# Arch / Manjaro
sudo pacman -S --needed bash curl jq fzf aria2 ca-certificates chafa

# Fedora
sudo dnf install -y bash curl jq fzf aria2 ca-certificates chafa

# openSUSE
sudo zypper install -y bash curl jq fzf aria2 ca-certificates chafa
```

```bash
mkdir -p ~/.local/bin
curl -fsSL https://gitee.com/spark-store-project/spark-store-tui/raw/master/package-root/usr/bin/spark-store-tui -o ~/.local/bin/spark-store-tui
chmod +x ~/.local/bin/spark-store-tui
export PATH="$HOME/.local/bin:$PATH"
MODE=apm spark-store-tui
```

非 Debian 系统下，`MODE=auto` 会默认进入 APM Store 模式。建议非 Debian 用户显式使用 `MODE=apm`，避免展示 Spark Store 的 `.deb` 软件包内容。

## 运行

```bash
spark-store-tui
```

只下载不安装，并保留下载目录：

```bash
INSTALL_AFTER_DOWNLOAD=0 KEEP_DOWNLOADS=1 spark-store-tui
```

## 环境变量

| 变量 | 说明 |
|---|---|
| `MODE` | 可选 `auto`、`spark`、`apm`、`choose`，默认 `auto` |
| `ARCH_PATH` | 覆盖自动检测的商店路径，例如 `amd64-store`、`arm64-store`、`amd64-apm`、`arm64-apm` |
| `STRICT_SPEC` | 为 `1` 时只使用文档化的 Spark/APM 元数据路径，默认 `1` |
| `DIRECT_FALLBACK` | 为 `1` 时 Metalink 下载失败后尝试直连包地址 |
| `INSTALL_AFTER_DOWNLOAD` | 可选 `auto`、`1`、`0`，设为 `0` 时只下载不安装 |
| `KEEP_DOWNLOADS` | 为 `1` 时保留下载目录并在退出时打印路径 |
| `DOWNLOAD_DIR` | 自定义下载目录；未设置时使用 `/tmp/spark-store-tui.xxxxxx` |
| `IMAGE_PREVIEW` | 为 `1` 时启用终端图片预览 |
| `IMAGE_KIND` | 可选 `auto`、`icon`、`screen`，默认 `auto` |
| `IMAGE_WIDTH` | chafa 或 viu 使用的图片预览宽度 |
| `IMAGE_HEIGHT` | chafa 使用的图片预览高度 |
| `ARIA2_READOUT` | 为 `0` 时关闭 aria2c 控制台输出；默认关闭 aria2c 摘要 |
| `ALLOW_SPARK_ON_NON_DEB` | 为 `1` 时允许在非 Debian 系统上测试 Spark Store 模式 |

## Spark Store 与 APM Store

Spark Store 模式使用 Spark 元数据，例如：

```text
https://d.spark-app.store/amd64-store/<category>/<pkgname>/app.json
```

启用安装时，如果系统存在 `ssinstall`，下载的包会通过 `sudo ssinstall <local_file>` 处理。

APM Store 模式使用 APM 元数据，例如：

```text
https://d.spark-app.store/amd64-apm/<category>/<pkgname>/app.json
```

本地包处理使用 `sudo apm ssaudit <local_file>`，并会显示在线安装提示：

```bash
sudo apm install -y <pkgname>
```

## 许可证

本项目使用 GPL-3.0-only 许可证发布，完整许可证文本见 `COPYING`。

## 灵感来源与友链

- 灵感来源：https://github.com/SHORiN-KiWATA/shorin-contrib/blob/main/pacman/pac
- 友链：我的偶像 SHORiN-KiWATA，https://github.com/SHORiN-KiWATA
- 星火应用商店官方 GitHub 仓库：https://github.com/spark-store-project/spark-store
- 星火应用商店官网：https://www.spark-app.store/

## 项目信息

- 作者：Xynrin
- 维护者：Xynrin <xynrin@163.com>
- Gitee 组织仓库：https://gitee.com/spark-store-project/spark-store-tui
- GitHub 备用源码：https://github.com/Xynrin/spark-store-tui
