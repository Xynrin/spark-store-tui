# 重构架构（第一期）

新实现位于 Go 模块中，现有 Bash 入口和既有 DEB/RPM/AUR 打包文件暂不修改。

```text
cmd/spark-store-tui        程序入口
internal/domain            与远程来源无关的应用、镜像和下载任务模型
internal/provider          Spark Store metadata 适配器
internal/download          aptss/aria2 进程适配和可恢复下载任务
internal/state             XDG 状态目录中的持久化下载任务
internal/ui                Bubble Tea 三栏交互界面
internal/system            发行版、CPU 架构和可选依赖的检测与补给命令
```

## 下载职责

1. Provider 将 Spark Store metadata 规范化为 `domain.App`，只负责展示信息和真实 Deb 文件名。
2. Debian 下载器从文件名提取真实包名并执行 `aptss download <包名>=<版本>`；aptss/aria2 负责软件源、镜像、摘要校验和续传。
3. TUI 记录外围任务状态，并结合最终文件与 `.aria2` 控制文件恢复中断任务。
4. 安装、卸载、版本查询和所选应用更新分别调用 aptss；APM 平台调用对应的 `apm` 命令。

## 运行开发版

需要 Go 1.25 或更新版本：

```bash
go run ./cmd/spark-store-tui
go test ./...
```

默认 TUI 仅绑定 Spark Store metadata：进入应用浏览时会先加载官方分类，再按选择的架构和分类加载应用列表。

启动时会自动根据 `/etc/os-release` 和 Go 运行时架构选择 `amd64-store` 或 `arm64-store`。终端图片预览使用 `chafa`；在仓库根目录运行时，首页会渲染 `icon.png`，也可使用 `SPARK_STORE_TUI_BRAND_IMAGE` 指定图片。不会在启动时自行安装任何软件，用户明确执行 `sparkstore --bootstrap-images` 后才会按发行版调用对应包管理器。
