# 0.8.0 原生包发布

发布版本统一使用 `0.8.0`，tag 为 `v0.8.0`。

| 渠道 | 产物 |
|---|---|
| GitHub / Gitee Release | `spark-store-tui-source-0.8.0.tar.gz`、每个架构的 `.deb`、RPM / SRPM |
| Debian | `spark-store-tui_0.8.0-1_amd64.deb` 或 `..._arm64.deb` |
| RPM | `spark-store-tui-0.8.0-1.<arch>.rpm` |
| AUR | `spark-store-tui`，从 GitHub source tarball 构建 |

## Debian

在目标架构的 Linux 环境中：

```bash
make check
make build
make source
```

上传生成的 `.deb` 和 source tarball 到 GitHub Release 与 Gitee Release。`make build` 会将打包根目录复制到 Linux 临时目录，以兼容 WSL `/mnt/c` 的权限模型。

## RPM

将源码包复制到 RPM 构建目录，然后构建：

```bash
mkdir -p ~/rpmbuild/SOURCES ~/rpmbuild/SPECS
cp spark-store-tui-source-0.8.0.tar.gz ~/rpmbuild/SOURCES/
cp packaging/rpm/spark-store-tui.spec ~/rpmbuild/SPECS/
rpmbuild -ba ~/rpmbuild/SPECS/spark-store-tui.spec
```

RPM 模板会使用 Go 构建本机架构二进制，并提供 `sparkstore` 和兼容命令链接。

## AUR

生成 source tarball 后，计算 SHA-256 并写入 `packaging/aur/PKGBUILD` 与 `.SRCINFO`。随后：

```bash
makepkg --printsrcinfo > .SRCINFO
git clone ssh://aur@aur.archlinux.org/spark-store-tui.git
cp PKGBUILD .SRCINFO spark-store-tui/
cd spark-store-tui
git add PKGBUILD .SRCINFO
git commit -m "spark-store-tui 0.8.0"
git push
```

只有拥有 AUR SSH 推送权限的维护者才能完成最后一步。
