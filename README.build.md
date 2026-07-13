# 构建 spark-store-tui 0.8.3

在 Linux 或 WSL（Go 1.25+、`dpkg-deb`）中：

```bash
go test ./...
go vet ./...
make build
sudo apt install ./spark-store-tui_0.8.3-2_$(dpkg --print-architecture).deb
```

`make build` 会在 Linux 临时目录中生成原生架构 `.deb`，因此仓库位于 WSL 的 `/mnt/c` 时也不会受 NTFS 权限限制。

生成的 Deb 声明 `aptss | spark-store` 为安装后端依赖。直接执行 `apt install` 前需已安装官方 `aptss` 或完整的 Spark Store；面向普通用户时推荐使用 `README.md` 中的一键安装命令，它会自动下载并校验 `aptss`。
