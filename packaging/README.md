# Native Package Manager Publishing

This directory contains packaging templates for package managers that can
install dependencies automatically.

## Arch Linux / AUR

`yay` searches the Arch repositories and AUR, not ArchWiki. To make this
package searchable with `yay -Ss spark-store-tui`, publish `packaging/aur` to
the AUR package repository named `spark-store-tui`.

Expected user command after AUR publication:

```bash
yay -S spark-store-tui
```

Maintainer publishing outline:

```bash
git clone ssh://aur@aur.archlinux.org/spark-store-tui.git
cp packaging/aur/PKGBUILD packaging/aur/.SRCINFO spark-store-tui/
cd spark-store-tui
git add PKGBUILD .SRCINFO
git commit -m "add spark-store-tui"
git push
```

Regenerate `.SRCINFO` whenever `PKGBUILD` metadata changes:

```bash
makepkg --printsrcinfo > .SRCINFO
```

## Fedora / COPR

For Fedora, the shortest practical path before official Fedora packaging is a
COPR repository.

Expected user commands after COPR publication:

```bash
sudo dnf copr enable xynrin/spark-store-tui
sudo dnf install spark-store-tui
```

Fedora Atomic, Silverblue and Kinoite users can enable the same COPR, then layer
the package:

```bash
sudo dnf copr enable xynrin/spark-store-tui
sudo rpm-ostree install spark-store-tui
systemctl reboot
```

Maintainer publishing outline:

```bash
mkdir -p ~/rpmbuild/SOURCES ~/rpmbuild/SPECS
cp spark-store-tui-deb-source-0.7.2.tar.gz ~/rpmbuild/SOURCES/
cp packaging/rpm/spark-store-tui.spec ~/rpmbuild/SPECS/
rpmbuild -bs ~/rpmbuild/SPECS/spark-store-tui.spec
copr-cli build xynrin/spark-store-tui ~/rpmbuild/SRPMS/spark-store-tui-0.7.2-1*.src.rpm
```

