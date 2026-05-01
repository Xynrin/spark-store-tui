# Native Package Manager Publishing

This directory contains packaging templates for package managers that can install dependencies automatically.

## Arch Linux / AUR

`yay` searches the Arch repositories and AUR, not ArchWiki. The package is published at:

```text
https://aur.archlinux.org/packages/spark-store-tui
```

User command:

```bash
yay -S spark-store-tui
```

Maintainer publishing outline:

```bash
git clone ssh://aur@aur.archlinux.org/spark-store-tui.git
cp packaging/aur/PKGBUILD packaging/aur/.SRCINFO spark-store-tui/
cd spark-store-tui
git add PKGBUILD .SRCINFO
git commit -m "update spark-store-tui"
git push
```

Regenerate `.SRCINFO` whenever `PKGBUILD` metadata changes:

```bash
makepkg --printsrcinfo > .SRCINFO
```

## Fedora / DNF / RPM

For Fedora, this repository can be used in two ways:

- a self-hosted RPM repository under GitHub Pages or the Gitee organization mirror
- a COPR repository for a more native Fedora workflow

Self-hosted RPM repository commands:

```bash
sudo tee /etc/yum.repos.d/spark-store-tui.repo >/dev/null <<'EOF'
[spark-store-tui]
name=Spark Store TUI
baseurl=https://xynrin.github.io/spark-store-tui/rpm
enabled=1
gpgcheck=0
repo_gpgcheck=0
EOF
sudo dnf install spark-store-tui
```

Domestic Gitee mirror:

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

Fedora Atomic, Silverblue and Kinoite users can use the same repository, then layer the package:

```bash
sudo rpm-ostree install spark-store-tui
systemctl reboot
```

Expected user commands after COPR publication:

```bash
sudo dnf copr enable xynrin/spark-store-tui
sudo dnf install spark-store-tui
```

Maintainer publishing outline:

```bash
mkdir -p ~/rpmbuild/SOURCES ~/rpmbuild/SPECS
cp spark-store-tui-deb-source-0.7.2.tar.gz ~/rpmbuild/SOURCES/
cp packaging/rpm/spark-store-tui.spec ~/rpmbuild/SPECS/
rpmbuild -bs ~/rpmbuild/SPECS/spark-store-tui.spec
copr-cli build xynrin/spark-store-tui ~/rpmbuild/SRPMS/spark-store-tui-0.7.2-1*.src.rpm
```
