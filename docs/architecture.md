# 重构架构（第一期）

新实现位于 Go 模块中，现有 Bash 入口和既有 DEB/RPM/AUR 打包文件暂不修改。

```text
cmd/spark-store-tui        程序入口
internal/domain            与远程来源无关的应用、镜像和下载任务模型
internal/provider          Spark Store metadata 适配器
internal/download          支持断点续传、校验和、原子落盘的下载器
internal/state             XDG 状态目录中的持久化下载任务
internal/ui                Bubble Tea 三栏交互界面
internal/system            发行版、CPU 架构和可选依赖的检测与补给命令
```

## 下载职责

1. Provider 将各平台的 API、metadata 或 release 数据规范化为 `domain.App`。
2. 镜像策略根据区域、健康检查和优先级选择 `domain.Mirror`。
3. 下载器写入 `*.part`，使用 HTTP Range 续传，并在 SHA-256 验证通过后原子改名。
4. 安装器将在第二期作为独立适配层加入；下载与提权安装不会再绑定为同一步。

## 运行开发版

需要 Go 1.25 或更新版本：

```bash
go run ./cmd/spark-store-tui
go test ./...
```

默认 TUI 仅绑定 Spark Store metadata：进入应用浏览时会先加载官方分类，再按选择的架构和分类加载应用列表。

启动时会自动根据 `/etc/os-release` 和 Go 运行时架构选择 `amd64-store` 或 `arm64-store`。终端图片预览使用 `chafa`；在仓库根目录运行时，首页会渲染 `icon.png`，也可使用 `SPARK_STORE_TUI_BRAND_IMAGE` 指定图片。不会在启动时自行安装任何软件，用户明确执行 `sparkstore --bootstrap-images` 后才会按发行版调用对应包管理器。
