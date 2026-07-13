#!/usr/bin/env bash
set -euo pipefail

root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
workdir=$(mktemp -d)
trap 'rm -rf "$workdir"' EXIT

git clone ssh://aur@aur.archlinux.org/spark-store-tui.git "$workdir/spark-store-tui"
install -m 0644 "$root/packaging/aur/PKGBUILD" "$workdir/spark-store-tui/PKGBUILD"
install -m 0644 "$root/packaging/aur/.SRCINFO" "$workdir/spark-store-tui/.SRCINFO"

cd "$workdir/spark-store-tui"
if git diff --quiet; then
  echo 'AUR package is already current.'
  exit 0
fi
git add PKGBUILD .SRCINFO
git commit -m 'spark-store-tui 0.8.0'
git push
