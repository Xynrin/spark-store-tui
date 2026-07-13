# 构建 spark-store-tui 0.8.3

在 Linux 或 WSL（Go 1.25+、`dpkg-deb`）中：

```bash
go test ./...
go vet ./...
make build
sudo apt install ./spark-store-tui_0.8.3-1_$(dpkg --print-architecture).deb
```

`make build` 会在 Linux 临时目录中生成原生架构 `.deb`，因此仓库位于 WSL 的 `/mnt/c` 时也不会受 NTFS 权限限制。
