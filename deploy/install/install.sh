#!/usr/bin/env sh
set -eu

# NexCtl Agent：从 GitHub Releases 下载二进制并写入配置（参考 nezhahq/scripts/agent 的体验优化）
# 仓库与产物：https://github.com/nexctl/agent （zip: nexctl_<goos>_<goarch>.zip）
#
# 安装（管道传参须加 -s --）：
#   curl -fsSL https://你的域名/deploy/install/install.sh | sh -s -- "http://控制面:8080" "节点令牌"
# 可选第 3 个参数：版本标签，如 v0.1.0（默认 latest）
# 卸载：
#   curl -fsSL .../install.sh | sh -s -- uninstall
#   或：sh install.sh uninstall
#
# 环境变量：
#   NEXCTL_AGENT_REPO   默认 nexctl/agent
#   CN=1                视为中国大陆网络，走 GitHub 代理加速下载（见 NEXCTL_GH_PROXY）
#   NEXCTL_NO_MIRROR=1  禁用代理，始终直连 GitHub
#   NEXCTL_GH_PROXY     默认 https://ghproxy.net/https://
#   NO_COLOR=1          禁用彩色输出

# ---------- 终端颜色（对齐 nezha 风格）----------
if [ -z "${NO_COLOR:-}" ] && [ -t 1 ]; then
  red=$(printf '\033[0;31m')
  green=$(printf '\033[0;32m')
  yellow=$(printf '\033[0;33m')
  plain=$(printf '\033[0m')
else
  red=''
  green=''
  yellow=''
  plain=''
fi

err() { printf '%s%s%s\n' "$red" "$*" "$plain" >&2; }
success() { printf '%s%s%s\n' "$green" "$*" "$plain"; }
info() { printf '%s%s%s\n' "$yellow" "$*" "$plain"; }

# ---------- sudo 封装（无 root 时尝试 sudo）----------
sudo_run() {
  _uid=$(id -ru)
  if [ "$_uid" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
      command sudo "$@"
    else
      err "需要 root 权限或已安装 sudo，无法继续。"
      exit 1
    fi
  else
    "$@"
  fi
}

deps_check() {
  _missing=""
  for dep in unzip grep; do
    if ! command -v "$dep" >/dev/null 2>&1; then
      _missing="${_missing} ${dep}"
    fi
  done
  if ! command -v curl >/dev/null 2>&1 && ! command -v wget >/dev/null 2>&1; then
    _missing="${_missing} curl|wget"
  fi
  if [ -n "$_missing" ]; then
    err "缺少依赖:${_missing}，请安装后重试。"
    exit 1
  fi
}

# ---------- 架构 / 系统（与 goreleaser 产物名对齐，参考 nezha env_check）----------
env_check() {
  mach=$(uname -m 2>/dev/null || echo "")
  case "$mach" in
    amd64 | x86_64) GOARCH=amd64 ;;
    i386 | i686) GOARCH=386 ;;
    aarch64 | arm64) GOARCH=arm64 ;;
    *arm*) GOARCH=arm ;;
    s390x) GOARCH=s390x ;;
    riscv64) GOARCH=riscv64 ;;
    mips) GOARCH=mips ;;
    mipsel | mipsle) GOARCH=mipsle ;;
    loongarch64) GOARCH=loong64 ;;
    *)
      err "未知架构: $mach"
      exit 1
      ;;
  esac

  system=$(uname -s 2>/dev/null || echo "")
  case "$system" in
    Linux*) GOOS=linux ;;
    Darwin*) GOOS=darwin ;;
    FreeBSD*) GOOS=freebsd ;;
    MINGW* | MSYS* | CYGWIN*) GOOS=windows ;;
    *)
      err "未知系统: $system"
      exit 1
      ;;
  esac
}

# ---------- 中国大陆网络：可选走 ghproxy（无官方 gitee 镜像时的常见做法）----------
geo_check() {
  isCN=""
  _ua="Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0"
  _text=""
  for _url in "https://www.cloudflare.com/cdn-cgi/trace" "https://cloudflare.com/cdn-cgi/trace"; do
    _text=$(curl -A "$_ua" -m 10 -sS "$_url" 2>/dev/null || true)
    if echo "$_text" | grep -q 'loc=CN'; then
      isCN=true
      break
    fi
  done
}

apply_github_mirror() {
  _url="$1"
  if [ -z "${isCN:-}" ] || [ -n "${NEXCTL_NO_MIRROR:-}" ]; then
    printf '%s' "$_url"
    return
  fi
  case "$_url" in
    https://github.com/*)
      _prefix="${NEXCTL_GH_PROXY:-https://ghproxy.net/https://}"
      printf '%s%s' "$_prefix" "${_url#https://}"
      ;;
    *)
      printf '%s' "$_url"
      ;;
  esac
}

download_file() {
  _url="$1"
  _dest="$2"
  if command -v curl >/dev/null 2>&1; then
    curl --max-time 60 -fsSL "$_url" -o "$_dest" && return 0
  fi
  if command -v wget >/dev/null 2>&1; then
    if wget --help 2>&1 | grep -q -- '--timeout'; then
      wget --timeout=60 -q -O "$_dest" "$_url" && return 0
    else
      { wget -T 60 -q -O "$_dest" "$_url" || wget -q -O "$_dest" "$_url"; } && return 0
    fi
  fi
  return 1
}

download_with_retry() {
  _url="$1"
  _dest="$2"
  _n=0
  while [ "$_n" -lt 3 ]; do
    _n=$((_n + 1))
    if download_file "$_url" "$_dest"; then
      return 0
    fi
    info "下载失败，重试 $_n/3 ..."
    sleep 2
  done
  err "下载失败，请检查网络或设置 NEXCTL_NO_MIRROR=1 后换网络重试。"
  return 1
}

# ---------- 卸载 ----------
uninstall_nexctl() {
  info "正在停止 nexctl-agent 进程..."
  if command -v pkill >/dev/null 2>&1; then
    pkill -f '[n]exctl-agent' 2>/dev/null || true
  fi
  sleep 1

  if [ -d /opt/nexctl ]; then
    info "移除 /opt/nexctl ..."
    sudo_run rm -rf /opt/nexctl
  fi
  if [ -f /usr/local/bin/nexctl-agent ]; then
    info "移除旧版 /usr/local/bin/nexctl-agent ..."
    sudo_run rm -f /usr/local/bin/nexctl-agent
  fi
  success "卸载步骤已执行（用户目录下的 ~/.config/nexctl、~/.local 等需自行清理）。"
}

# ---------- 主流程：卸载 ----------
if [ "${1:-}" = "uninstall" ]; then
  uninstall_nexctl
  exit 0
fi

SERVER_URL="${1:-}"
ENROLLMENT_TOKEN="${2:-}"
RELEASE_TAG="${3:-}"

GITHUB_REPO="${NEXCTL_AGENT_REPO:-nexctl/agent}"

if [ -z "$SERVER_URL" ] || [ -z "$ENROLLMENT_TOKEN" ]; then
  err "用法: curl -fsSL .../install.sh | sh -s -- <服务器地址> <节点token> [版本标签]"
  err "卸载: curl -fsSL .../install.sh | sh -s -- uninstall"
  exit 1
fi

deps_check
env_check

if [ "$GOOS" = windows ] && [ "$GOARCH" = arm ]; then
  err "当前仓库未提供 Windows ARM 构建。"
  exit 1
fi

# CN：可手动 export CN=1，否则自动探测（与 nezha 一致）
if [ "${CN:-}" = "1" ] || [ "${CN:-}" = "true" ]; then
  isCN=true
  info "已设置 CN=1，将使用 GitHub 下载代理（NEXCTL_NO_MIRROR=1 可禁用）。"
else
  geo_check
  if [ -n "${isCN:-}" ]; then
    isCN=true
    info "检测到可能位于中国大陆网络，将使用 GitHub 下载代理（设置 NEXCTL_NO_MIRROR=1 可禁用）。"
  fi
fi

# ---------- 下载 URL ----------
ZIP_NAME="nexctl_${GOOS}_${GOARCH}.zip"
if [ -z "$RELEASE_TAG" ]; then
  _zip_url="https://github.com/${GITHUB_REPO}/releases/latest/download/${ZIP_NAME}"
else
  TAG="$RELEASE_TAG"
  case "$TAG" in
    v*) ;;
    *) TAG="v${TAG}" ;;
  esac
  _zip_url="https://github.com/${GITHUB_REPO}/releases/download/${TAG}/${ZIP_NAME}"
fi
ZIP_URL=$(apply_github_mirror "$_zip_url")

# ---------- 安装路径（root：/opt/nexctl，与 nezha /opt/nezha 类似）----------
if [ "$(id -u)" -ne 0 ] && [ -z "${HOME:-}" ]; then
  err "非 root 安装需要已设置 HOME。"
  exit 1
fi

if [ "$(id -u)" -eq 0 ]; then
  NZ_BASE="/opt/nexctl"
  BIN_DIR="${NZ_BASE}/agent"
  CONFIG_DIR="${NZ_BASE}/agent"
  DATA_ROOT="${NZ_BASE}"
else
  BIN_DIR="${HOME:-}/.local/bin"
  CONFIG_DIR="${XDG_CONFIG_HOME:-${HOME:-}/.config}/nexctl"
  DATA_ROOT="${XDG_DATA_HOME:-${HOME:-}/.local/share}/nexctl"
fi

CONFIG_FILE="${CONFIG_DIR}/agent.yaml"

NODE_NAME=$(hostname -s 2>/dev/null || hostname 2>/dev/null || echo "nexctl-node")

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

info "正在下载: ${ZIP_URL}"
if ! download_with_retry "$ZIP_URL" "${TMPDIR}/${ZIP_NAME}"; then
  exit 1
fi

unzip -q -o "${TMPDIR}/${ZIP_NAME}" -d "$TMPDIR"

BIN_NAME=""
for _name in nexctl-agent nexctl-agent.exe; do
  if [ -f "${TMPDIR}/${_name}" ]; then
    BIN_NAME="$_name"
    break
  fi
done
if [ -z "$BIN_NAME" ]; then
  err "压缩包内未找到 nexctl-agent:"
  ls -la "$TMPDIR" >&2
  exit 1
fi

BIN_PATH="${BIN_DIR}/${BIN_NAME}"

mkdir -p "$BIN_DIR" "$CONFIG_DIR" "${DATA_ROOT}/data/config" "${DATA_ROOT}/data/credentials" "${DATA_ROOT}/data/logs"

if command -v install >/dev/null 2>&1; then
  install -m 0755 "${TMPDIR}/${BIN_NAME}" "$BIN_PATH"
else
  cp "${TMPDIR}/${BIN_NAME}" "$BIN_PATH"
  chmod 0755 "$BIN_PATH"
fi

yaml_escape() {
  printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g'
}

tee "$CONFIG_FILE" >/dev/null <<EOF
agent:
  server_url: "$(yaml_escape "$SERVER_URL")"
  enrollment_token: "$(yaml_escape "$ENROLLMENT_TOKEN")"
  node_name: "$(yaml_escape "$NODE_NAME")"
  data_dir: "${DATA_ROOT}/data"
  config_dir: "${DATA_ROOT}/data/config"
  credential_dir: "${DATA_ROOT}/data/credentials"
  log_dir: "${DATA_ROOT}/data/logs"
  github_repo: "$(yaml_escape "$GITHUB_REPO")"
  disable_auto_update: false
  self_update_period_minutes: 0
EOF
chmod 0644 "$CONFIG_FILE"

# 标记由本脚本安装（便于日后扩展）
if [ "$(id -u)" -eq 0 ] && [ -d /opt/nexctl/agent ]; then
  echo "nexctl-install-script" > /opt/nexctl/agent/.nexctl-install
fi

STDOUT_LOG="${DATA_ROOT}/data/logs/install-stdout.log"
# nohup 需对非 root 可写路径；root 下 /opt/nexctl 已可写
nohup "$BIN_PATH" -config "$CONFIG_FILE" >>"$STDOUT_LOG" 2>&1 &
AGENT_PID=$!

sleep 1
if ! kill -0 "$AGENT_PID" 2>/dev/null; then
  err "警告: agent 进程可能已异常退出，请查看 ${STDOUT_LOG} 与 ${DATA_ROOT}/data/logs/agent.log"
fi

echo ""
success "安装完成。"
echo "  二进制: ${BIN_PATH}"
echo "  配置: ${CONFIG_FILE}"
info "  已在后台启动 agent（PID ${AGENT_PID}），日志目录: ${DATA_ROOT}/data/logs"
echo "  标准输出/错误追加: ${STDOUT_LOG}"
echo ""
echo "如需前台调试:"
echo "  kill ${AGENT_PID}"
echo "  ${BIN_PATH} -config ${CONFIG_FILE}"
echo ""

if [ "$(id -u)" -ne 0 ]; then
  case ":${PATH}:" in
    *:"${BIN_DIR}":*) ;;
    *)
      info "若找不到命令，可将目录加入 PATH:"
      echo "  export PATH=\"${BIN_DIR}:\$PATH\""
      echo ""
      ;;
  esac
fi
