<p align="center">
  <img src="icon.png" width="180" alt="Spark Store TUI logo">
</p>

<h1 align="center">Spark Store TUI</h1>

<p align="center">星火终端助手 · 原生 Linux 软件管理终端界面</p>

<p align="center">
  <a href="https://gitee.com/spark-store-project">
    <img src="https://foruda.gitee.com/avatar/1758374023182638862/6258778_spark-store-project_1758374023.png!avatar100" height="44" alt="星火商店 Spark Store">
    <br>
    <strong>本项目附属于星火商店</strong>
  </a>
</p>

<p align="center">
  <a href="https://gitee.com/spark-store-project/spark-store-tui/releases/tag/v0.8.3"><img src="https://img.shields.io/badge/Release-v0.8.3-ff9d2e" alt="Release v0.8.3"></a>
  <a href="https://gitee.com/spark-store-project/spark-store-tui"><img src="https://img.shields.io/badge/Gitee-Source-C71D23?logo=gitee&logoColor=white" alt="Gitee source"></a>
  <a href="COPYING"><img src="https://img.shields.io/badge/License-GPL--3.0--only-4caf50" alt="GPL-3.0-only"></a>
  <img src="https://img.shields.io/badge/Go-%3E%3D1.25-00ADD8?logo=go&logoColor=white" alt="Go 1.25 or newer">
</p>

> `sparkstore` 使用 Spark Store 公开 metadata 展示目录；Debian 应用统一交给官方 `aptss` 下载、校验、安装、更新和卸载，其他受支持发行版交给 Amber APM。

当前稳定版：**0.8.3**；Debian 包修订版：**0.8.3-3**。

## 功能

- 自动识别 Debian/Ubuntu、Arch、Fedora/RHEL 与 openSUSE，以及 `x86_64` / `aarch64` / `loongarch64` 架构。
- 分类浏览、搜索、图标终端预览（可选 `chafa`）和长简介展开。
- Debian 使用 `aptss/aria2` 完成镜像下载、软件源摘要校验与断点续传，并持久化外围任务状态。
- 应用重启后自动识别完整包和 `.aria2` 中断任务；中断任务按 `D` 即可继续。
- 选中应用按 `P` 查询已安装版本与候选版本，只升级该应用；支持卸载和可选删除本地安装包。

应用来源只保留 **Spark Store**，不再显示第三方 Release 软件源或 APM Store。

## 依赖与下载流程

目录展示和 `/` 搜索使用 TUI 已读取的 Spark Store metadata，不会为每次按键启动外部进程；Debian 的实际软件包操作全部依赖 **`aptss | spark-store`**。`ca-certificates` 也是必需依赖；`sudo` 与用于图片预览的 `chafa` 是推荐依赖。

Debian/Ubuntu/UOS/麒麟上的正常流程如下：

1. 根据系统架构和分类读取 Spark Store 公开元数据。
2. 从元数据中的 Deb 文件名取得真实包名，例如目录 `vscode` 对应 `code_…_amd64.deb`，实际包名是 `code`；同时核对架构，拒绝下载其他 CPU 的包。
3. 按 `D` 后在 `~/Downloads` 执行 `aptss download <包名>=<版本>`。`aptss/aria2` 使用星火软件源配置选择镜像、验证仓库摘要并保存续传状态；TUI 会临时退出全屏界面以显示原生下载进度。
4. 外围任务状态保存在 `~/.config/sparkstore/tasks.json`。进程意外退出后，再次启动会结合最终文件和 `.aria2` 控制文件识别完成或中断状态。
5. 下载完成后按 `I` 或 `Enter`，执行 `aptss install <本地.deb> -y`，依赖解析和桌面文件安装均由软件包及 aptss 负责；按 `Esc` 只保留安装包。
6. 选中应用按 `P` 会执行 `aptss policy <包名>`。发现候选版本高于已安装版本时再次按 `P` 或 `Enter`，执行 `aptss install --only-upgrade <包名> -y`，不会顺带升级全部系统软件。

如果 `~/Downloads` 已有同名完整包，再按 `D` 会提示直接安装、重新下载、删除或取消，不会静默重复下载。一键安装脚本会从星火官方软件包服务器下载、校验并安装 `aptss`；如果直接安装 Release 中的 Deb，则需要本机已经安装 `aptss` 或完整的 Spark Store。

可在终端单独运行 `aptss search spark-` 检查星火软件源是否已经初始化；TUI 的目录搜索仍以 metadata 为准，下载时才调用 aptss 的真实包索引。若软件源尚未初始化，请执行 `sudo aptss ssupdate`。

Arch、Fedora/RHEL 与 openSUSE 不在主机上直接安装 Spark Store 的 `.deb`，选中应用后会通过 Amber APM 安装；`P` 使用 `apm policy` 与 `apm install --only-upgrade` 查询和更新所选应用。缺少 `apm` 时会给出错误。一键安装脚本会尝试按发行版补齐 Amber APM，但 TUI 启动本身不依赖它。

## 命令

标准命令为：

```bash
sparkstore
```

安装包另外提供 `SparkStore`、`SPARKSTORE` 和旧命令 `spark-store-tui`。Linux 命令名区分大小写，未列出的大小写组合不能由 shell 自动识别。

## WSL / 本地开发

需要 Go 1.25 或更新版本。在 WSL 的 `/mnt/c` 目录中可直接执行：

```bash
go test ./...
go vet ./...
mkdir -p build
go build -buildvcs=false -o build/sparkstore ./cmd/spark-store-tui
./build/sparkstore
```

`-buildvcs=false` 避免 Git 对 Windows 挂载目录的所有权检查。图片预览是可选的：运行 `sparkstore --bootstrap-images` 可按发行版安装 `chafa`。

## 快捷键

进入应用页后：`/` 搜索、`[` / `]` 切换分类、`D` 下载/续传、`I` 安装、`P` 检查/更新所选应用、`U` 卸载、`E` 展开简介、`R` 刷新、`q` 退出。

下载中断或进程被关闭后，下一次启动会将任务标记为“可继续”，而不会卡住 UI；选中对应应用后按 `D` 继续下载。

## 发布物与后缀

| 平台 | 发布物 | 后缀 / 架构 |
|---|---|---|
| Debian / Ubuntu / UOS / 麒麟 | Debian 包 | `spark-store-tui_0.8.3-3_amd64.deb`、`..._arm64.deb`、`..._loong64.deb` |
| Fedora / RHEL / openSUSE / 银河麒麟 | RPM | `spark-store-tui-0.8.3-4.<arch>.rpm`（含 `loongarch64`） |
| Arch Linux | AUR 源包 | `spark-store-tui`（构建本机架构二进制） |
| 通用构建 | 源码包 | `spark-store-tui-source-0.8.3-r4.tar.gz` |

`v0.8.3` 同时提供 `amd64` / `x86_64`、`arm64` / `aarch64` 与 `loong64` / `loongarch64` 的 Deb、RPM。仓库中的 [`apt/`](apt/README.md) 与 [`rpm/`](rpm/README.md) 只是冻结在 `0.7.2` 的历史归档，为避免已有软件源立即失效而保留，不代表当前版本，也不再用于新安装或更新；Gitee 当前没有 APT/RPM Pages 仓库。请使用本页的一键安装命令或 `v0.8.3` Release，并只下载与本机架构匹配的包。

### CPU 架构选择

安装前先执行 `uname -m`，并按输出选择软件包：

| `uname -m` | Deb 架构 | RPM 架构 |
|---|---|---|
| `x86_64` | `amd64` | `x86_64` |
| `aarch64` | `arm64` | `aarch64` |
| `loongarch64` | `loong64` | `loongarch64` |

不要在架构不匹配时安装 amd64 包；其他架构请优先从源码构建。

### Debian / Ubuntu

推荐使用后文的一键安装命令，它会自动补齐官方 `aptss`。如果需要直接安装发布物，可同时安装经过固定 SHA-256 校验的官方 `aptss`：

```bash
curl -LO https://d.spark-app.store/store/depends/aptss_4.8.1-1_all.deb
echo 'cd95de3488f7e39ce0300b1e3ba38b0c9416871e68fb91098011ace26f057751  aptss_4.8.1-1_all.deb' | sha256sum -c -
curl -LO https://gitee.com/spark-store-project/spark-store-tui/releases/download/v0.8.3/spark-store-tui_0.8.3-3_amd64.deb
sudo apt install ./aptss_4.8.1-1_all.deb ./spark-store-tui_0.8.3-3_amd64.deb
sudo aptss ssupdate
```

aarch64 设备将文件名中的 `amd64` 替换为 `arm64`；龙芯设备替换为 `loong64`。

### 一键安装（国内镜像）

安装器会检测发行版和 CPU 架构；Debian 系会先补齐并校验官方 `aptss`，Deb/RPM 再下载并校验对应发行包。缺少本机架构附件时，它会明确询问是否从同一 tag 的源码构建，不会静默安装其他架构。Arch 仍按 AUR 规则调用 `yay` 安装 TUI；Arch/Fedora 上的星火应用通过 Amber APM 安装。

```bash
bash <(curl -fsSL https://gitee.com/spark-store-project/spark-store-tui/raw/master/scripts/install-sparkstore.sh) --mirror gitee
```

命令已固定使用国内镜像，不再显示镜像选择提示。直接运行脚本而不附带参数时，仍可交互选择网络环境。

### 国内网络 / Gitee 安装

Debian / Ubuntu 已安装 `aptss` 或 Spark Store 时，可直接下载 Gitee Release：

```bash
curl -LO https://gitee.com/spark-store-project/spark-store-tui/releases/download/v0.8.3/spark-store-tui_0.8.3-3_amd64.deb
sudo apt install ./spark-store-tui_0.8.3-3_amd64.deb
```

Gitee 也可用于源码构建：

```bash
git clone --depth 1 --branch v0.8.3 https://gitee.com/spark-store-project/spark-store-tui.git
cd spark-store-tui
go build -buildvcs=false -o sparkstore ./cmd/spark-store-tui
sudo install -Dm755 sparkstore /usr/local/bin/sparkstore
sparkstore
```

RPM 直链：

```text
https://gitee.com/spark-store-project/spark-store-tui/releases/download/v0.8.3/spark-store-tui-0.8.3-4.x86_64.rpm
```

### RPM 系统

下载对应发行版与架构的 `.rpm` 后执行：

```bash
sudo dnf install ./spark-store-tui-0.8.3-4.x86_64.rpm
```

openSUSE 可使用：

```bash
sudo zypper install ./spark-store-tui-0.8.3-4.x86_64.rpm
```

### Arch Linux / AUR

已发布的 AUR Git 仓库与当前 Release 保持相同版本。AUR RPC/yay 索引在刚推送后可能短暂缓存旧版本；若 `yay -Si spark-store-tui` 仍显示旧版本，请等待索引同步，或直接从 AUR Git 仓库构建：

```bash
sudo pacman -S --needed base-devel go git
git clone https://aur.archlinux.org/spark-store-tui.git
cd spark-store-tui
makepkg -si
```

索引同步后可简化为 `yay -S spark-store-tui`。AUR 会在本机架构构建，因此适合 `aarch64` 与 `loongarch64` 等没有预编译附件的平台。

## 构建发布物

在对应 Linux 构建环境中：

```bash
make check
make build       # 本机架构 .deb
make source      # 通用 source tarball
```

RPM 与 AUR 模板位于 [packaging](packaging/README.md)。

## 许可证

GPL-3.0-only，详见 [COPYING](COPYING)。
