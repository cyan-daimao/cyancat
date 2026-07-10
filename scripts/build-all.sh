#!/usr/bin/env bash
# build-all.sh — 一条命令交叉编译 cyancat 到多个平台
#
# 用法：
#   ./scripts/build-all.sh                          # 构建默认平台
#   ./scripts/build-all.sh windows/amd64            # 只构建指定平台
#   ./scripts/build-all.sh darwin/arm64 windows/amd64 linux/amd64
#
# 产物归档到 build/dist/，文件名带平台后缀，避免互相覆盖。
#
# 依赖（macOS 宿主交叉编译）：
#   brew install go node                            # 基础工具链
#   brew install mingw-w64                          # Windows 交叉编译（提供 x86_64-w64-mingw32-gcc）
#   go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
#   # Linux 交叉（可选，仅构建 linux/* 时需要）：
#   brew install FiloSottile/musl-cross/musl-cross
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# ────────────────────────────────────────────────────────────
# 默认目标平台（win11 amd + macOS arm64/intel 用户最常用）
# ────────────────────────────────────────────────────────────
DEFAULT_TARGETS=(
  "darwin/arm64"
  "darwin/amd64"
  "windows/amd64"
)

# 命令行传入平台则用传入的，否则用默认
if [ "$#" -gt 0 ]; then
  TARGETS=("$@")
else
  TARGETS=("${DEFAULT_TARGETS[@]}")
fi

# 平台 → C 交叉编译器（macOS 宿主，本机 clang 可交叉 darwin amd64）
declare -A CC_FOR=(
  ["darwin/arm64"]=""
  ["darwin/amd64"]=""
  ["darwin/universal"]=""
  ["windows/amd64"]="x86_64-w64-mingw32-gcc"
  ["windows/arm64"]="aarch64-w64-mingw32-gcc"
  ["linux/amd64"]="x86_64-linux-musl-gcc"
  ["linux/arm64"]="aarch64-linux-musl-gcc"
)
declare -A CXX_FOR=(
  ["windows/amd64"]="x86_64-w64-mingw32-g++"
  ["windows/arm64"]="aarch64-w64-mingw32-g++"
  ["linux/amd64"]="x86_64-linux-musl-g++"
  ["linux/arm64"]="aarch64-linux-musl-g++"
)

DIST="$ROOT/build/dist"
BIN="$ROOT/build/bin"

# ────────────────────────────────────────────────────────────
# 依赖检查
# ────────────────────────────────────────────────────────────
check_tool() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "❌ 缺少依赖：$1"
    case "$1" in
      wails)               echo "   安装：go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0" ;;
      x86_64-w64-mingw32-gcc) echo "   安装：brew install mingw-w64" ;;
      x86_64-linux-musl-gcc)  echo "   安装：brew install FiloSottile/musl-cross/musl-cross" ;;
      aarch64-w64-mingw32-gcc) echo "   安装：brew install mingw-w64（arm64 mingw 需单独启用）" ;;
      *) ;;
    esac
    exit 1
  }
}

# 检测当前 macOS 版本是否 >= 15.4
is_macos_15_4_or_later() {
  if [[ "$(uname -s)" != "Darwin" ]]; then
    return 1
  fi
  local version
  version=$(sw_vers -productVersion 2>/dev/null || echo "0.0")
  local IFS=.
  local -a current=($version)
  local -a required=(15 4 0)
  for i in 0 1 2; do
    local c="${current[$i]:-0}"
    local r="${required[$i]}"
    if (( c > r )); then
      return 0
    elif (( c < r )); then
      return 1
    fi
  done
  return 0
}

echo "🔧 检查基础依赖..."
check_tool go
check_tool wails
check_tool npm

# 预检所有目标的交叉编译器
for t in "${TARGETS[@]}"; do
  cc="${CC_FOR[$t]:-}"
  if [ -n "$cc" ]; then
    check_tool "$cc"
  fi
done

# 预生成 bindings（仅 macOS 15.4+ 需要）
# Wails 交叉编译 Windows/Linux 时，生成 bindings 阶段仍在本机 darwin 上跑 CGO，
# 会和 SDK 的 strchrnul 声明冲突；先在本机生成好 bindings，后续 build 跳过此阶段。
if is_macos_15_4_or_later; then
  echo "🔧 预生成 Wails bindings（macOS 15.4+ 需要 CGO_CFLAGS=-DHAVE_STRCHRNUL）..."
  CGO_ENABLED=1 CGO_CFLAGS="-DHAVE_STRCHRNUL" wails generate module
fi

# 清空归档目录
rm -rf "$DIST"
mkdir -p "$DIST"

# ────────────────────────────────────────────────────────────
# 单平台构建
# ────────────────────────────────────────────────────────────

build_one() {
  local target="$1"
  local cc="${CC_FOR[$target]:-}"
  local cxx="${CXX_FOR[$target]:-}"
  local cgo_cflags=""
  local stem="${target//\//-}"   # darwin/arm64 -> darwin-arm64

  echo ""
  echo "════════════════════════════════════════"
  echo "📦 构建 $target"
  echo "════════════════════════════════════════"

  if [ -n "$cc" ]; then
    # 跨平台：指定 C/C++ 工具链
    echo "   CC  = $cc"
    echo "   CXX = $cxx"
  else
    # 本机/同族：darwin 用 clang，CGO 开启（go-sqlite3 需要）
    cc=clang
    cxx=clang++
    unset GOOS GOARCH  # wails build -platform 会设
  fi

  # macOS 15.4+ 上 SDK 声明了 strchrnul 但带可用性检查，需要显式设置 HAVE_STRCHRNUL
  # 避免 "static declaration of 'strchrnul' follows non-static declaration"。
  # 旧版 macOS / 其他平台没有 strchrnul，不能盲目设置该 flag。
  case "$target" in
    darwin/*)
      if is_macos_15_4_or_later; then
        cgo_cflags="-DHAVE_STRCHRNUL"
        echo "   CGO_CFLAGS = $cgo_cflags"
      fi
      ;;
  esac

  # 清掉上次残留产物，避免归档时误判
  rm -f "$BIN/cyancat" "$BIN/cyancat.exe"
  rm -rf "$BIN/cyancat.app"

  # wails build 自动跑前端 npm build + 后端 go build
  # 环境变量仅作用于本次构建，避免 darwin 的 CGO_CFLAGS 泄漏给 windows。
  # bindings 已在脚本开头预生成，所有平台均跳过，避免 windows 交叉编译时
  # 在 darwin 主机上再次触发 CGO 编译。
  CC="$cc" CXX="$cxx" CGO_ENABLED=1 CGO_CFLAGS="$cgo_cflags" \
    wails build -platform "$target" -trimpath -clean -skipbindings

  # 归档产物到 build/dist
  # macOS：app 名称保持 cyancat.app，压缩包统一命名为 cyancat-mac.zip
  case "$target" in
    darwin/*)
      if [ -d "$BIN/cyancat.app" ]; then
        mkdir -p "$DIST"
        cp -R "$BIN/cyancat.app" "$DIST/cyancat.app"
        (cd "$DIST" && zip -qr "cyancat-mac.zip" "cyancat.app")
        echo "   ✅ $DIST/cyancat-mac.zip (app 名称保持 cyancat.app)"
      else
        echo "   ⚠️  未找到 cyancat.app"
      fi
      ;;
    windows/*)
      if [ -f "$BIN/cyancat.exe" ]; then
        mv "$BIN/cyancat.exe" "$DIST/cyancat-$stem.exe"
        (cd "$DIST" && zip -qj "cyancat-$stem.zip" "cyancat-$stem.exe")
        echo "   ✅ $DIST/cyancat-$stem.exe (+ .zip)"
      else
        echo "   ⚠️  未找到 cyancat.exe"
      fi
      ;;
    linux/*)
      if [ -f "$BIN/cyancat" ]; then
        mv "$BIN/cyancat" "$DIST/cyancat-$stem"
        (cd "$DIST" && tar czf "cyancat-$stem.tar.gz" "cyancat-$stem")
        echo "   ✅ $DIST/cyancat-$stem (+ .tar.gz)"
      else
        echo "   ⚠️  未找到 cyancat"
      fi
      ;;
    *)
      echo "   ⚠️  未知平台：$target"
      ;;
  esac
}

# ────────────────────────────────────────────────────────────
# 逐平台构建
# ────────────────────────────────────────────────────────────
FAILED=()
for t in "${TARGETS[@]}"; do
  if ! build_one "$t"; then
    FAILED+=("$t")
  fi
done

# ────────────────────────────────────────────────────────────
# 汇总
# ────────────────────────────────────────────────────────────
echo ""
echo "════════════════════════════════════════"
echo "🎉 构建完成，产物目录：$DIST"
echo "════════════════════════════════════════"
if [ ${#FAILED[@]} -gt 0 ]; then
  echo "❌ 失败的平台：${FAILED[*]}"
fi
ls -lh "$DIST"
