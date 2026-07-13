# Spark Store TUI / 星火终端助手

`sparkstore` 是面向 Linux 的原生终端应用管理器。它只读取 Spark Store 的公开 metadata，通过官方 Metalink 选择下载镜像，并在明确确认后调用当前发行版的包管理器安装或卸载本地软件包。

当前版本：`0.8.0`。

## 功能

- 自动识别 Debian/Ubuntu、Arch、Fedora/RHEL 与 openSUSE，以及 `x86_64` / `aarch64` 架构。
- 分类浏览、搜索、图标终端预览（可选 `chafa`）和长简介展开。
- Metalink 镜像回退、断点续传、下载任务持久化与无数据超时检测。
- 应用重启后自动识别完成包和中断任务；中断任务按 `D` 即可继续。
- `.deb`、RPM 和 Arch 本地包的确认安装；原生包管理器卸载；可选删除本地安装包。

应用来源只保留 **Spark Store**，不再显示 GitHub Releases、Gitee Releases 或 APM Store。

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

进入应用页后：`/` 搜索、`[` / `]` 切换分类、`D` 下载/续传、`I` 安装、`U` 卸载、`E` 展开简介、`R` 刷新、`q` 退出。

下载中断或进程被关闭后，下一次启动会将任务标记为“可继续”，而不会卡住 UI；选中对应应用后按 `D` 继续下载。

## 发布物与后缀

| 平台 | 发布物 | 后缀 / 架构 |
|---|---|---|
| Debian / Ubuntu | Debian 包 | `spark-store-tui_0.8.0-1_amd64.deb`、`..._arm64.deb` |
| Fedora / RHEL / openSUSE | RPM | `spark-store-tui-0.8.0-1.<arch>.rpm` |
| Arch Linux | AUR 源包 | `spark-store-tui`（构建本机架构二进制） |
| 通用构建 | 源码包 | `spark-store-tui-source-0.8.0.tar.gz` |

GitHub Release 是主发布渠道；Gitee 同步相同 tag、源码与二进制发布物。请只下载与本机架构匹配的包。

### Debian / Ubuntu

```bash
curl -LO https://github.com/Xynrin/spark-store-tui/releases/download/v0.8.0/spark-store-tui_0.8.0-1_amd64.deb
sudo apt install ./spark-store-tui_0.8.0-1_amd64.deb
```

arm64 设备将文件名中的 `amd64` 替换为 `arm64`。

### RPM 系统

下载对应发行版与架构的 `.rpm` 后执行：

```bash
sudo dnf install ./spark-store-tui-0.8.0-1.x86_64.rpm
```

openSUSE 可使用：

```bash
sudo zypper install ./spark-store-tui-0.8.0-1.x86_64.rpm
```

### Arch Linux / AUR

发布完成后：

```bash
yay -S spark-store-tui
```

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
