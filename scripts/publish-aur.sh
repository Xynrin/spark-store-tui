#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT

git clone ssh://aur@aur.archlinux.org/spark-store-tui.git "$workdir/spark-store-tui"
install -m 0644 "$root/packaging/aur/PKGBUILD" "$workdir/spark-store-tui/PKGBUILD"
install -m 0644 "$root/packaging/aur/.SRCINFO" "$workdir/spark-store-tui/.SRCINFO"

cd "$workdir/spark-store-tui"
git config user.name "${GIT_AUTHOR_NAME:-Xynrin}"
git config user.email "${GIT_AUTHOR_EMAIL:-xynrin@163.com}"
if git diff --quiet; then
  echo 'AUR package is already current.'
  exit 0
fi
git add PKGBUILD .SRCINFO
version=$(awk -F= '/^pkgver=/ { print $2; exit }' PKGBUILD)
git commit -m "spark-store-tui $version"
git push
