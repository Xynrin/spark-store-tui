# spark-store-tui v0.8.1

- 重构为 Go 原生 TUI，应用目录收敛为 Spark Store。
- 下载任务持久化；重启后识别完成包和中断 `.part`，按 `D` 可继续。
- 下载传输无数据超时后会失败并允许重试，避免无限卡住。
- root 身份运行时不再调用 `sudo`，修复 Ubuntu `sudo-rs` 的 DBUS 环境误解析问题。
- Release 同时提供 Deb `amd64` / `arm64` 与 RPM `x86_64` / `aarch64`；AUR 从源码构建本机二进制。
- 标准启动命令为 `sparkstore`；保留 `SparkStore`、`SPARKSTORE`、`spark-store-tui` 兼容入口。
