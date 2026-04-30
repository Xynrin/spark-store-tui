# spark-store-tui v0.7.2

- 新增：非 Debian 系默认只显示 APM，不显示 Spark。
- 新增：根据 `/etc/os-release` 自动选择 Spark/APM。
- 新增：根据 `uname -m` 自动选择 amd64/arm64 架构路径。
- 改进：图片预览优先使用 `app.json` 图标字段，减少错图。
- 改进：chafa 图片渲染保持比例，减少失真。
- 改进：下载默认进入 `/tmp`，退出自动清理。
- 改进：`KEEP_DOWNLOADS=1` 可保留下载目录。
- 改进：aria2c 默认不刷屏。
- 许可证：GPL-3.0-only。
