# vNext 状态与已知限制

## 已完成

- Spark Store 分类、应用 metadata、图标字段的真实读取。
- 应用来源已收敛为 Spark Store；不再包含 GitHub/Gitee Release Provider。
- Debian 应用通过 aptss/aria2 下载、校验和续传，并持久化外围任务记录。
- Debian 通过 aptss、Arch/RPM/openSUSE 通过 Amber APM 安装、卸载及更新所选应用；TUI 需要明确确认才会执行。
- 按 `P` 查询已安装版本与候选版本，只升级当前选中应用。
- 已下载安装包的检测：避免重复下载，允许安装、重下、删除或取消。
- 搜索、分类切换、窄终端布局、图片加载防抖和本地图片缓存。

## 尚未完成

- TUI 内还没有下载队列、暂停、取消和历史页；Debian 下载时会显示 aptss/aria2 的原生实时进度。
- AppImage、Flatpak、Spark `.spk` 与 APM 审计安装还没有统一安装实现；当前自动安装覆盖 `.deb`、`.rpm` 和 Arch 本地包。
- 应用详情目前显示图标；截图图库、截图切换和懒加载尚未完成。
- Go vNext 二进制尚未替换现有 DEB/RPM/AUR 的 Bash 包装与发布流程。

## 已知问题

- 首次缓存远程图片仍会受镜像网络影响，超过 8 秒会放弃显示；后续命中缓存会很快。
- 很窄的终端会启用单栏摘要模式，图片和完整详情会隐藏以保证可操作性。
- 已安装状态依赖 aptss/APM 的包数据库；AppImage 和 Flatpak 还没有统一识别。
