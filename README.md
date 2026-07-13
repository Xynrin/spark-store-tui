# Spark Store TUI 0.7.2 兼容归档

此 APT/RPM 仓库已停止维护，仅为避免旧用户的软件源更新报错而保留。不要再使用这里的 `0.7.2-1` 安装新系统。

迁移到当前版本：

```bash
sudo rm -f /etc/apt/sources.list.d/spark-store-tui.list
sudo apt update
bash <(curl -fsSL https://raw.githubusercontent.com/Xynrin/spark-store-tui/main/scripts/install-sparkstore.sh)
```

国内网络：

```bash
bash <(curl -fsSL https://gitee.com/spark-store-project/spark-store-tui/raw/master/scripts/install-sparkstore.sh) --mirror gitee
```

当前 Release：<https://github.com/Xynrin/spark-store-tui/releases/tag/v0.8.3>

旧仓库签名密钥指纹：`1AE6D4E7C4DB8C016F72F8C6A4D276F9CF8E57A9`
