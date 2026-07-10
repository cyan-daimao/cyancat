# cyancat 打包指南（macOS + Windows）

本项目目标平台为 **macOS** 与 **Windows**，两者都可在一台 macOS（Apple Silicon）上打出。

---

## 平台能力矩阵

| 目标平台 | 在 macOS 本机构建 | 工具链 |
|----------|:----------------:|--------|
| darwin/arm64 | ✅ 原生 | Xcode CLT |
| darwin/amd64 | ✅ 交叉（同 OS 跨架构） | Xcode CLT |
| darwin/universal | ✅ 合并 arm64+amd64（推荐分发） | Xcode CLT |
| windows/amd64 | ✅ 交叉编译 | mingw-w64 |

---

## 前置依赖

```bash
# Wails CLI（本项目 v2.12.0）
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0

# macOS 构建
xcode-select --install

# Windows 交叉编译工具链（国内网络建议用镜像源，见下）
brew install mingw-w64
```

**国内网络装 mingw-w64**：Homebrew 的 bottle 走 ghcr.io，国内常超时。用中科大镜像源：

```bash
HOMEBREW_BOTTLE_DOMAIN=https://mirrors.ustc.edu.cn/homebrew-bottles \
  brew install mingw-w64
```

自检：`wails doctor` / `which x86_64-w64-mingw32-gcc`

---

## macOS 包

```bash
wails build -platform darwin/arm64 -clean        # Apple Silicon
wails build -platform darwin/amd64 -clean        # Intel
wails build -platform darwin/universal -clean    # 通用二进制（一个 .app 同时支持两种芯片）
```

产物：`build/bin/cyancat.app`

验证 universal 是否含双架构：

```bash
lipo -info build/bin/cyancat.app/Contents/MacOS/cyancat
# → Architectures in the fat file: ... are: x86_64 arm64
```

---

## Windows 包

Windows 版需要 CGO，交叉编译依赖 mingw-w64 提供的 `x86_64-w64-mingw32-gcc`：

```bash
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
  wails build -platform windows/amd64 -clean
```

产物：`build/bin/cyancat.exe`

---

## 一键脚本

`scripts/build-all.sh` 默认打 **mac(arm64/amd64) + windows(amd64)**，并把产物按平台后缀
归档到 `build/dist/`（带 zip）：

```bash
./scripts/build-all.sh                      # 默认三平台
./scripts/build-all.sh darwin/universal     # 只打指定平台
./scripts/build-all.sh windows/amd64
```

---

## 产物位置

| 平台 | `build/bin/` | `build-all.sh` 归档到 `build/dist/` |
|------|--------------|-------------------------------------|
| macOS | `cyancat.app` | `cyancat-darwin-<arch>.app` + `.zip` |
| Windows | `cyancat.exe` | `cyancat-windows-amd64.exe` + `.zip` |

---

## 构建实测结果

环境：macOS (Apple Silicon) + wails v2.12.0 + go1.25，2026-07-01 实测全部通过。

| 平台 | 状态 | 备注 |
|------|:----:|------|
| darwin/arm64 | ✅ | 原生，10.8s，产物 `cyancat.app` |
| darwin/amd64 | ✅ | 含于 universal 的 x86_64 分片（`lipo -info` 确认） |
| darwin/universal | ✅ | 44s，fat 二进制含 `x86_64 + arm64` |
| windows/amd64 | ✅ | mingw-w64 (GCC 16.1.0) 交叉编译，10.8s，产物 `cyancat.exe` |

---

## 常见问题

- **windows 报 CGO/gcc 找不到**：没装或没指定 mingw，确认 `brew install mingw-w64` 且构建时带
  `CC=x86_64-w64-mingw32-gcc`。
- **想要 CI 自动出包**：GitHub Actions 用 `macos-latest` runner，一个 job 即可打 mac + windows
  （job 内 `brew install mingw-w64` 后按上面命令构建）。

## macOS 15.4+ 编译注意事项

macOS 15.4 及更新版本在 SDK 中声明了 `strchrnul`，但默认带可用性检查，导致 `pg_query_go` / `go-sqlite3` 的 CGO 编译报 "static declaration of 'strchrnul' follows non-static declaration"。

`scripts/build-all.sh` 已自动检测主机版本，仅在 macOS 15.4+ 构建 darwin 目标时设置：

```bash
export CGO_CFLAGS="-DHAVE_STRCHRNUL"
```

手动构建时，**只有**在 macOS 15.4+ 且遇到上述错误时才设置：

```bash
export CGO_CFLAGS="-DHAVE_STRCHRNUL"
wails build -platform darwin/arm64 -clean
```

> ⚠️ 旧版 macOS、Linux 或 Windows 交叉编译环境不要设置 `-DHAVE_STRCHRNUL`，否则会出现 `implicit declaration of function 'strchrnul'` 错误。
