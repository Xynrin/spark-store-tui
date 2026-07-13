#!/usr/bin/env bash
# One-command installer for Spark Store TUI. It accepts a public GitHub or
# Gitee release mirror and never silently installs an older AUR package.
set -euo pipefail

RELEASE_VERSION="${SPARKSTORE_VERSION:-0.8.3}"
MIRROR="${SPARKSTORE_MIRROR:-}"
OWNER="Xynrin"
REPOSITORY="spark-store-tui"
GITEE_OWNER="spark-store-project"
TEMP_DIR=""

cleanup() { [ -z "$TEMP_DIR" ] || rm -rf "$TEMP_DIR"; }
trap cleanup EXIT

usage() {
  cat <<'EOF'
Usage: install-sparkstore.sh [--mirror github|gitee]

Environment:
  SPARKSTORE_VERSION  Version without v prefix (default: 0.8.3)
  SPARKSTORE_MIRROR   github or gitee
EOF
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --mirror) MIRROR="${2:-}"; shift 2 ;;
    --help|-h) usage; exit 0 ;;
    *) echo "Unknown argument: $1" >&2; usage >&2; exit 2 ;;
  esac
done

if [ "$(uname -s)" != Linux ]; then
  echo '此安装器仅支持 Linux。' >&2
  exit 1
fi

SUDO=()
if [ "$(id -u)" -ne 0 ]; then
  command -v sudo >/dev/null || { echo '需要 sudo 或 root 权限。' >&2; exit 1; }
  SUDO=(sudo)
fi

source /etc/os-release 2>/dev/null || true
identity="${ID:-} ${ID_LIKE:-}"
case "${identity,,}" in
  *arch*|*manjaro*|*endeavouros*) FAMILY="arch" ;;
  *debian*|*ubuntu*|*deepin*|*uos*|*mint*) FAMILY="deb" ;;
  *suse*) FAMILY="suse" ;;
  *fedora*|*rhel*|*centos*|*rocky*|*alma*) FAMILY="rpm" ;;
  *) echo "不支持的发行版：${PRETTY_NAME:-unknown}" >&2; exit 1 ;;
esac

case "$(uname -m)" in
  x86_64|amd64) DEB_ARCH="amd64"; RPM_ARCH="x86_64" ;;
  aarch64|arm64) DEB_ARCH="arm64"; RPM_ARCH="aarch64" ;;
  loongarch64) DEB_ARCH="loong64"; RPM_ARCH="loongarch64" ;;
  riscv64) DEB_ARCH="riscv64"; RPM_ARCH="riscv64" ;;
  *) echo "不支持的 CPU 架构：$(uname -m)" >&2; exit 1 ;;
esac

# chafa enables the in-terminal logo and screenshot previews. It is optional
# for the binary itself, so a missing package repository must not prevent the
# main application from being installed.
ensure_image_preview() {
  command -v chafa >/dev/null 2>&1 && return 0
  echo '正在补齐终端图片预览依赖 chafa…'
  case "$FAMILY" in
    arch)
      if ! yay -S --needed chafa; then
        echo '提示：chafa 安装失败，星火商店仍可使用；可稍后手动安装以启用图片预览。' >&2
      fi
      ;;
    deb)
      if ! "${SUDO[@]}" apt-get update || ! "${SUDO[@]}" apt-get install -y chafa; then
        echo '提示：chafa 安装失败，星火商店仍可使用；可稍后手动安装以启用图片预览。' >&2
      fi
      ;;
    rpm)
      if ! "${SUDO[@]}" dnf install -y chafa; then
        echo '提示：chafa 安装失败，星火商店仍可使用；可稍后手动安装以启用图片预览。' >&2
      fi
      ;;
    suse)
      if ! "${SUDO[@]}" zypper --non-interactive install chafa; then
        echo '提示：chafa 安装失败，星火商店仍可使用；可稍后手动安装以启用图片预览。' >&2
      fi
      ;;
  esac
}

# Amber APM is the compatibility layer used by Spark Store applications on
# Arch and Fedora. The commands follow Amber PM's published installation path.
ensure_amber_runtime() {
  command -v apm >/dev/null 2>&1 && return 0
  echo '正在补齐 Amber APM 运行环境…'
  case "$FAMILY" in
    arch)
      if ! yay -S --needed --noconfirm amber-package-manager; then
        echo '提示：Amber APM 安装失败；TUI 已安装，但 Arch 应用安装功能暂不可用。' >&2
      fi
      ;;
    rpm)
      if ! "${SUDO[@]}" dnf -y copr enable xmp360/spark-store || ! "${SUDO[@]}" dnf install -y spark-store; then
        echo '提示：Amber APM 安装失败；TUI 已安装，但 Fedora 应用安装功能暂不可用。' >&2
      fi
      ;;
  esac
}

if [ "$FAMILY" = arch ]; then
  command -v yay >/dev/null || { echo 'Arch 系统请先安装 yay。' >&2; exit 1; }
  aur_version=$(yay -Si spark-store-tui 2>/dev/null | awk -F: '/^Version[[:space:]]*:/ {gsub(/[[:space:]]/, "", $2); print $2; exit}')
  case "$aur_version" in
    "$RELEASE_VERSION"-*|"$RELEASE_VERSION") ;;
    *)
      # The AUR RPC index can lag behind a successful Git push. Check the
      # published PKGBUILD without executing it before rejecting the install.
      aur_probe=$(mktemp -d)
      aur_pkgver=""
      if git clone -q --depth 1 https://aur.archlinux.org/spark-store-tui.git "$aur_probe/repository"; then
        aur_pkgver=$(awk -F= '/^pkgver=/ {gsub(/[[:space:]]/, "", $2); print $2; exit}' "$aur_probe/repository/PKGBUILD")
      fi
      rm -rf "$aur_probe"
      if [ "$aur_pkgver" != "$RELEASE_VERSION" ]; then
        echo "AUR 当前版本为 ${aur_version:-未找到}，尚未发布 $RELEASE_VERSION；为避免安装旧版已停止。" >&2
        exit 1
      fi
      echo "AUR API 尚未刷新（当前 ${aur_version:-未知}），已从 AUR Git 确认 $RELEASE_VERSION。"
      ;;
  esac
  yay -S --needed --noconfirm spark-store-tui
  ensure_image_preview
  ensure_amber_runtime
  exit 0
fi

if [ -z "$MIRROR" ]; then
  printf '选择下载网络 [1] GitHub  [2] Gitee： '
  read -r choice
  case "$choice" in
    1|github|GitHub) MIRROR="github" ;;
    2|gitee|Gitee) MIRROR="gitee" ;;
    *) echo '请输入 1 或 2。' >&2; exit 2 ;;
  esac
fi
case "$MIRROR" in github|gitee) ;; *) echo '镜像只能是 github 或 gitee。' >&2; exit 2 ;; esac

ensure_installer_tools() {
  command -v awk >/dev/null 2>&1 && return 0
  echo '正在补齐安装器校验依赖 gawk…'
  case "$FAMILY" in
    arch) yay -S --needed gawk ;;
    deb) "${SUDO[@]}" apt-get update; "${SUDO[@]}" apt-get install -y gawk ;;
    rpm) "${SUDO[@]}" dnf install -y gawk ;;
    suse) "${SUDO[@]}" zypper --non-interactive install gawk ;;
  esac
  command -v awk >/dev/null 2>&1 || { echo '未能安装 gawk，无法校验下载文件。' >&2; exit 1; }
}

ensure_installer_tools

TEMP_DIR=$(mktemp -d)
# apt downloads local packages as the _apt user when possible.  Allow it to
# traverse this temporary directory so installation stays sandboxed.
chmod 755 "$TEMP_DIR"

download() {
  destination="$1"; shift
  for url in "$@"; do
    rm -f "$destination"
    if curl --fail --location --retry 2 --connect-timeout 15 --output "$destination" "$url"; then
      return 0
    fi
  done
  return 1
}

verify() {
  file="$1" expected="$2"
  [ -n "$expected" ] || return 0
  actual=$(sha256sum "$file" | awk '{print $1}')
  [ "$actual" = "$expected" ] || { echo "SHA-256 校验失败：$file" >&2; exit 1; }
}

release_url() {
  asset="$1"
  if [ "$MIRROR" = github ]; then
    printf '%s\n' "https://github.com/$OWNER/$REPOSITORY/releases/download/v$RELEASE_VERSION/$asset"
  else
    # Gitee exposes a public, redirecting release URL.  Its attachment API
    # requires a release ID and may return 401 to anonymous installers.
    printf '%s\n' "https://gitee.com/$GITEE_OWNER/$REPOSITORY/releases/download/v$RELEASE_VERSION/$asset"
  fi
}

confirm_source_build() {
  echo "未能从 $MIRROR 下载 $1。"
  printf '是否改为从源码构建？这会安装 Go 和构建依赖 [y/N]： '
  read -r answer
  case "${answer,,}" in
    y|yes) install_source ;;
    *) echo '已取消；未修改系统。' >&2; exit 1 ;;
  esac
}

install_source() {
  echo "未找到 $MIRROR 的本机二进制附件，改为从同一 tag 构建。"
  case "$FAMILY" in
    deb) "${SUDO[@]}" apt-get update; "${SUDO[@]}" apt-get install -y git golang-go ca-certificates ;;
    rpm) "${SUDO[@]}" dnf install -y git golang ca-certificates ;;
    suse) "${SUDO[@]}" zypper --non-interactive install git go ca-certificates ;;
  esac
  command -v go >/dev/null || { echo '未能安装 Go。' >&2; exit 1; }
  go_version=$(go env GOVERSION | sed 's/^go//')
  if [ "$(printf '1.25\n%s\n' "$go_version" | sort -V | head -n1)" != 1.25 ]; then
    echo "需要 Go 1.25+，当前为 $go_version。" >&2
    exit 1
  fi
  repo_url="https://github.com/$OWNER/$REPOSITORY.git"
  [ "$MIRROR" = gitee ] && repo_url="https://gitee.com/$GITEE_OWNER/$REPOSITORY.git"
  source_ref="refs/tags/v$RELEASE_VERSION"
  # v0.8.3 predates the RPM/APM packaging revision. Pin its source fallback
  # to the reviewed r2 snapshot instead of rebuilding the old tag.
  [ "$RELEASE_VERSION" = '0.8.3' ] && source_ref='0fb4dd9774518c3f18f41a1a165332d304cd24ba'
  git init -q "$TEMP_DIR/source"
  git -C "$TEMP_DIR/source" remote add origin "$repo_url"
  git -C "$TEMP_DIR/source" fetch --depth 1 origin "$source_ref"
  git -C "$TEMP_DIR/source" checkout -q --detach FETCH_HEAD
  (
    cd "$TEMP_DIR/source"
    go build -buildvcs=false -o sparkstore ./cmd/spark-store-tui
    "${SUDO[@]}" install -Dm755 sparkstore /usr/local/bin/sparkstore
  )
  ensure_image_preview
  [ "$FAMILY" != rpm ] || ensure_amber_runtime
  echo '安装完成：/usr/local/bin/sparkstore'
}

case "$FAMILY" in
  deb)
    asset="spark-store-tui_${RELEASE_VERSION}-1_${DEB_ARCH}.deb"
    checksum=""
    [ "$asset" = 'spark-store-tui_0.8.0-1_amd64.deb' ] && checksum='f081a2ed817410c72f810ca5cc97cd2cbbb4b6bec19cf8c15152d2264515f738'
    [ "$asset" = 'spark-store-tui_0.8.1-1_amd64.deb' ] && checksum='447f3f2ad66d00b07a42e8c057553c9a1d12c1fab250fcc0c1c0f3df07b7ddad'
    [ "$asset" = 'spark-store-tui_0.8.1-1_arm64.deb' ] && checksum='094efc497d867e8d1e54de71bd1fc475001ca5dec90fa6332befff03317fafae'
    [ "$asset" = 'spark-store-tui_0.8.2-1_amd64.deb' ] && checksum='53e872153e807a4ef0a7787792cff0a2e245a76df7b83d65ca8ff8c753c1ce2d'
    [ "$asset" = 'spark-store-tui_0.8.2-1_arm64.deb' ] && checksum='cba638b0822f9bd052a7dea8209f6cba6e705b74d9878a087a854d3053836e0d'
    [ "$asset" = 'spark-store-tui_0.8.2-1_loong64.deb' ] && checksum='4d2d26d0e9bcbe4fa905f41a6d27c0f94e687833628106bd02e6cfda45cb2999'
    [ "$asset" = 'spark-store-tui_0.8.3-1_amd64.deb' ] && checksum='b46dbf0b0ec025e50c60d30e0e63b7ede7fa042f7603a990def0189b285ac943'
    [ "$asset" = 'spark-store-tui_0.8.3-1_arm64.deb' ] && checksum='22bdf415b07f804aa08b800682c816c20bc08dc7db75ff836c86758bfb85a510'
    [ "$asset" = 'spark-store-tui_0.8.3-1_loong64.deb' ] && checksum='13646a6640f272530850f84c21a29ccf2c1d159935c7fde8caad9f18ec7f2517'
    if download "$TEMP_DIR/$asset" "$(release_url "$asset")"; then
      verify "$TEMP_DIR/$asset" "$checksum"
      "${SUDO[@]}" apt-get install -y "$TEMP_DIR/$asset"
      ensure_image_preview
    else
      confirm_source_build "$asset"
    fi
    ;;
  rpm|suse)
    rpm_release=1
    [ "$RELEASE_VERSION" = '0.8.3' ] && rpm_release=2
    asset="spark-store-tui-${RELEASE_VERSION}-${rpm_release}.${RPM_ARCH}.rpm"
    checksum=""
    [ "$asset" = 'spark-store-tui-0.8.0-1.x86_64.rpm' ] && checksum='e7e230456ddb0581c0dc3b45d1a620aa3cfe634344ccaddfa285023a05a545be'
    [ "$asset" = 'spark-store-tui-0.8.1-1.x86_64.rpm' ] && checksum='837fdd1085d2a943e8f8f895ba56782939f49bb9d1ad307f1a4787aca5c3b30f'
    [ "$asset" = 'spark-store-tui-0.8.1-1.aarch64.rpm' ] && checksum='cd87e55e3883604aaf8edd28b77700be90ad1424b8462857e956d9afb293e47e'
    [ "$asset" = 'spark-store-tui-0.8.2-1.x86_64.rpm' ] && checksum='46ad1fd28dd32fdf91dbbb36f10a835086d45bd7dc7302594c250bf8bdcafeb4'
    [ "$asset" = 'spark-store-tui-0.8.2-1.aarch64.rpm' ] && checksum='fd6044d8c784733ee7d0d71d8aaee9a21d074dcd00dba8bf5f0842ad44d71348'
    [ "$asset" = 'spark-store-tui-0.8.2-1.loongarch64.rpm' ] && checksum='571c321af5bda5063f026336ef03cccc03bfd6d38cf466a829f7889b0f90b6bf'
    [ "$asset" = 'spark-store-tui-0.8.3-1.x86_64.rpm' ] && checksum='9ebc97d5dd57fb16acdf2b9de3ef46c5780af6cc19ccd1e40578bf5e152b6b77'
    [ "$asset" = 'spark-store-tui-0.8.3-1.aarch64.rpm' ] && checksum='b0d6cc7a24e158191ef97d1ff16e1dc211ead8309216d9b1b9cee048f0add9f1'
    [ "$asset" = 'spark-store-tui-0.8.3-1.loongarch64.rpm' ] && checksum='eb7687d525c63592f84f02500bad97acc945787149bc0cc127e02f57d2f6229b'
    [ "$asset" = 'spark-store-tui-0.8.3-2.x86_64.rpm' ] && checksum='18511e142a9ee050f905c749a6e148cd2988842bf4d94ccdce2079eabbab73b8'
    [ "$asset" = 'spark-store-tui-0.8.3-2.aarch64.rpm' ] && checksum='bf6d93faa75b17a3049586eab806117456baa8685558011cda4ff9e4667a4fea'
    [ "$asset" = 'spark-store-tui-0.8.3-2.loongarch64.rpm' ] && checksum='f490edb3904fe4f95ccd2cbffe5517335874c085c41c6e80ad00efddb93de92d'
    if download "$TEMP_DIR/$asset" "$(release_url "$asset")"; then
      verify "$TEMP_DIR/$asset" "$checksum"
      if [ "$FAMILY" = rpm ]; then
        "${SUDO[@]}" dnf install -y "$TEMP_DIR/$asset"
      else
        "${SUDO[@]}" zypper --non-interactive install "$TEMP_DIR/$asset"
      fi
      ensure_image_preview
      [ "$FAMILY" != rpm ] || ensure_amber_runtime
    else
      confirm_source_build "$asset"
    fi
    ;;
esac
