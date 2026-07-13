#!/usr/bin/env bash
# One-command installer for Spark Store TUI. It accepts a public GitHub or
# Gitee release mirror and never silently installs an older AUR package.
set -euo pipefail

RELEASE_VERSION="${SPARKSTORE_VERSION:-0.8.1}"
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
  SPARKSTORE_VERSION  Version without v prefix (default: 0.8.1)
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

if [ "$FAMILY" = arch ]; then
  command -v yay >/dev/null || { echo 'Arch 系统请先安装 yay。' >&2; exit 1; }
  aur_version=$(yay -Si spark-store-tui 2>/dev/null | awk -F: '/^Version[[:space:]]*:/ {gsub(/[[:space:]]/, "", $2); print $2; exit}')
  case "$aur_version" in "$RELEASE_VERSION"-*|"$RELEASE_VERSION") ;; *)
    echo "AUR 当前版本为 ${aur_version:-未找到}，尚未发布 $RELEASE_VERSION；为避免安装旧版已停止。" >&2
    exit 1
  esac
  yay -S --needed spark-store-tui
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
  git clone --depth 1 --branch "v$RELEASE_VERSION" "$repo_url" "$TEMP_DIR/source"
  (
    cd "$TEMP_DIR/source"
    go build -buildvcs=false -o sparkstore ./cmd/spark-store-tui
    "${SUDO[@]}" install -Dm755 sparkstore /usr/local/bin/sparkstore
  )
  echo '安装完成：/usr/local/bin/sparkstore'
}

case "$FAMILY" in
  deb)
    asset="spark-store-tui_${RELEASE_VERSION}-1_${DEB_ARCH}.deb"
    checksum=""
    [ "$asset" = 'spark-store-tui_0.8.0-1_amd64.deb' ] && checksum='f081a2ed817410c72f810ca5cc97cd2cbbb4b6bec19cf8c15152d2264515f738'
    [ "$asset" = 'spark-store-tui_0.8.1-1_amd64.deb' ] && checksum='447f3f2ad66d00b07a42e8c057553c9a1d12c1fab250fcc0c1c0f3df07b7ddad'
    [ "$asset" = 'spark-store-tui_0.8.1-1_arm64.deb' ] && checksum='094efc497d867e8d1e54de71bd1fc475001ca5dec90fa6332befff03317fafae'
    if download "$TEMP_DIR/$asset" "$(release_url "$asset")"; then
      verify "$TEMP_DIR/$asset" "$checksum"
      "${SUDO[@]}" apt-get install -y "$TEMP_DIR/$asset"
    else
      confirm_source_build "$asset"
    fi
    ;;
  rpm|suse)
    asset="spark-store-tui-${RELEASE_VERSION}-1.${RPM_ARCH}.rpm"
    checksum=""
    [ "$asset" = 'spark-store-tui-0.8.0-1.x86_64.rpm' ] && checksum='e7e230456ddb0581c0dc3b45d1a620aa3cfe634344ccaddfa285023a05a545be'
    [ "$asset" = 'spark-store-tui-0.8.1-1.x86_64.rpm' ] && checksum='837fdd1085d2a943e8f8f895ba56782939f49bb9d1ad307f1a4787aca5c3b30f'
    [ "$asset" = 'spark-store-tui-0.8.1-1.aarch64.rpm' ] && checksum='cd87e55e3883604aaf8edd28b77700be90ad1424b8462857e956d9afb293e47e'
    if download "$TEMP_DIR/$asset" "$(release_url "$asset")"; then
      verify "$TEMP_DIR/$asset" "$checksum"
      if [ "$FAMILY" = rpm ]; then
        "${SUDO[@]}" dnf install -y "$TEMP_DIR/$asset"
      else
        "${SUDO[@]}" zypper --non-interactive install "$TEMP_DIR/$asset"
      fi
    else
      confirm_source_build "$asset"
    fi
    ;;
esac
