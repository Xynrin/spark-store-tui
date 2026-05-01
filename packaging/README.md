# 原生包管理器发布说明

本目录保存 Arch/AUR、RPM/DNF 等原生包管理器的打包模板。目标是让用户可以像安装普通软件包一样自动拉取依赖并安装 `spark-store-tui`。

## Arch Linux / AUR

`yay` 搜索的是 Arch 官方仓库与 AUR，不是 ArchWiki。当前包已经发布到 AUR：

```text
https://aur.archlinux.org/packages/spark-store-tui
```

用户安装命令：

```bash
yay -S spark-store-tui
```

维护者发布流程：

```bash
git clone ssh://aur@aur.archlinux.org/spark-store-tui.git
cp packaging/aur/PKGBUILD packaging/aur/.SRCINFO spark-store-tui/
cd spark-store-tui
git add PKGBUILD .SRCINFO
git commit -m "update spark-store-tui"
git push
```

每次 `PKGBUILD` 元数据变化后都要重新生成 `.SRCINFO`：

```bash
makepkg --printsrcinfo > .SRCINFO
```

## Fedora / DNF / RPM

Fedora 当前可以使用本 Gitee 组织仓库中的自建 RPM 源：

```bash
sudo tee /etc/yum.repos.d/spark-store-tui.repo >/dev/null <<'EOF'
[spark-store-tui]
name=Spark Store TUI
baseurl=https://gitee.com/spark-store-project/spark-store-tui/raw/master/rpm
enabled=1
gpgcheck=0
repo_gpgcheck=0
EOF
sudo dnf install spark-store-tui
```

Fedora Atomic、Silverblue、Kinoite 可以使用同一个源，然后通过 rpm-ostree 分层安装：

```bash
sudo rpm-ostree install spark-store-tui
systemctl reboot
```

后续如果发布 COPR，可以切换为更原生的 Fedora 工作流：

```bash
sudo dnf copr enable xynrin/spark-store-tui
sudo dnf install spark-store-tui
```

维护者构建 SRPM 的参考流程：

```bash
mkdir -p ~/rpmbuild/SOURCES ~/rpmbuild/SPECS
cp spark-store-tui-deb-source-0.7.2.tar.gz ~/rpmbuild/SOURCES/
cp packaging/rpm/spark-store-tui.spec ~/rpmbuild/SPECS/
rpmbuild -bs ~/rpmbuild/SPECS/spark-store-tui.spec
copr-cli build xynrin/spark-store-tui ~/rpmbuild/SRPMS/spark-store-tui-0.7.2-1*.src.rpm
```
