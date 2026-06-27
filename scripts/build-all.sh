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
  local stem="${target//\//-}"   # darwin/arm64 -> darwin-arm64

  echo ""
  echo "════════════════════════════════════════"
  echo "📦 构建 $target"
  echo "════════════════════════════════════════"

  if [ -n "$cc" ]; then
    # 跨平台：指定 C/C++ 工具链
    export CC="$cc" CXX="$cxx" CGO_ENABLED=1
    echo "   CC  = $cc"
    echo "   CXX = $cxx"
  else
    # 本机/同族：darwin 用 clang，CGO 开启（go-sqlite3 需要）
    export CC=clang CXX=clang++ CGO_ENABLED=1
    unset GOOS GOARCH  # wails build -platform 会设
  fi

  # 清掉上次残留产物，避免归档时误判
  rm -f "$BIN/cyancat" "$BIN/cyancat.exe"
  rm -rf "$BIN/cyancat.app"

  # wails build 自动跑前端 npm build + 后端 go build
  wails build -platform "$target" -trimpath -clean

  # 归档产物到 build/dist，文件名带平台后缀
  case "$target" in
    darwin/*)
      if [ -d "$BIN/cyancat.app" ]; then
        mv "$BIN/cyancat.app" "$DIST/cyancat-$stem.app"
        (cd "$DIST" && zip -qr "cyancat-$stem.zip" "cyancat-$stem.app")
        echo "   ✅ $DIST/cyancat-$stem.app (+ .zip)"
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
