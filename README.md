# DBStudio (cyancat)

<p align="center">
  <strong>轻量级跨平台数据库桌面客户端</strong><br>
  面向开发者、数据工程师和轻量级 DBA 的 Navicat 风格数据库管理工具
</p>

---

## ✨ 特性

- 🔌 **多数据源管理** — MySQL / MariaDB / PostgreSQL，加密存储连接凭据
- 🌳 **对象树浏览器** — 数据库 → 表 → 字段 / 索引 / 外键，逐层懒加载
- 🎨 **可视化设计器** — 表设计器 + 字段网格，所有结构变更先预览 DDL 再执行
- 📝 **SQL 编辑器** — 基于 Monaco Editor，支持执行 / 格式化 / 注释切换
- 📊 **结果网格** — 虚拟滚动、列宽拖拽、分页、复制为 TSV / INSERT SQL / CSV / Markdown
- ⚠️ **安全确认** — 高风险 DDL（DROP TABLE / DROP DATABASE）强制二次确认
- 🔒 **凭据加密** — AES-GCM 加密存储密码，支持 OS Keychain 扩展
- 🖥️ **跨平台** — macOS / Windows / Linux

## 📸 截图

> *TODO: 添加应用截图*

## 🚀 快速开始

### 前置条件

| 依赖 | 版本要求 |
|------|---------|
| Go | 1.25+ |
| Node.js | 任何现代 LTS |
| Wails CLI | v2 `go install github.com/wailsapp/wails/v2/cmd/wails@latest` |
| macOS | WebKit（系统内置） |
| Windows | WebView2（Windows 10+ 内置） |
| Linux | WebKit2GTK (`sudo apt install libwebkit2gtk-4.0-dev`) |

### 开发模式

```bash
# 克隆仓库
git clone https://github.com/cyan-daimao/cyancat.git
cd cyancat

# 启动开发模式（后端热重载 + 前端热更新）
wails dev
```

### 构建

```bash
# 当前平台
wails build

# 指定平台
wails build -platform darwin/universal    # macOS 通用二进制
wails build -platform windows/amd64       # Windows
wails build -platform linux/amd64         # Linux
```

产物输出到 `build/bin/`。

### 常用命令

```bash
# 前端构建
cd frontend && npm run build

# Go 编译检查
go build ./...
go vet ./...
go test ./...

# 修改 Go API 签名后重新生成 Wails 前端绑定
wails generate module
```

## ⚙️ 运行时配置

| 配置项 | 说明 |
|--------|------|
| `CYANCAT_MASTER_KEY` | 64 位十六进制（32 字节）AES 主密钥，推荐生产环境设置 |
| `~/.cyancat/master.key` | 32 字节原始密钥文件（备选方案） |
| `~/.cyancat/cyancat.db` | 本地 SQLite 数据库（自动创建） |

> 开发模式未设置密钥时使用硬编码默认值，会在日志中输出警告。

## 🏗️ 技术架构

### 技术栈

| 层 | 技术 |
|----|------|
| **桌面框架** | Wails v2（Go + WebView） |
| **后端** | Go 1.25, GORM/SQLite, go-sql-driver/mysql, jackc/pgx |
| **前端** | React 18 + TypeScript + Vite |
| **样式** | Tailwind CSS + shadcn/ui（Radix 原语） |
| **状态管理** | Zustand |
| **编辑器** | Monaco Editor |
| **表格** | TanStack Table + TanStack Virtual |

### DDBD 四层架构

```
adapter  →  application  →  domain  →  infra
```

| 层 | 路径 | 职责 |
|----|------|------|
| **adapter** | `internal/adapter/` | Wails API 绑定、DTO 与转换函数 |
| **application** | `internal/application/<biz>service/` | 服务接口与实现、Cmd/Query/BO 对象 |
| **domain** | `internal/domain/<biz>/` | 充血模型实体、Repository 接口 |
| **infra** | `internal/infra/` | GORM 仓库实现、驱动抽象、会话管理、加密、日志 |

三个垂直切片：`connectionservice`、`queryservice`、`schemaservice`。

**转换链**（每层边界显式 `ToXxx` 函数，无反射）：
```
读取:  DO → Domain → BO → DTO
写入:  Request → Cmd → Domain → Repository → DO
```

### 项目结构

```
cyancat/
├── main.go                          # 入口 + DI 组装
├── wails.json                       # Wails 配置
├── internal/
│   ├── adapter/
│   │   ├── app.go                   # Wails App 结构体
│   │   ├── dto/                     # connection/query/schema DTO
│   │   └── http/                    # connection/query/schema API
│   ├── application/
│   │   ├── connectionservice/       # 连接管理
│   │   ├── queryservice/            # SQL 执行 + 历史
│   │   └── schemaservice/           # Schema 浏览 + DDL
│   ├── domain/
│   │   └── connection/              # 连接领域模型
│   └── infra/
│       ├── api/                     # Response[T] 统一响应
│       ├── crypto/                  # AES-GCM 加解密
│       ├── db/                      # SQLite 数据源 + GORM 仓库
│       ├── driver/                  # Driver/Conn/Dialect/Inspector 接口
│       │   ├── mysql/               # MySQL 驱动实现
│       │   └── postgres/            # PostgreSQL 驱动实现
│       ├── eventbus/                # Wails 事件总线
│       ├── keychain/                # OS Keychain（V1 AES 降级）
│       ├── logger/                  # zerolog 日志
│       └── session/                 # 运行时连接池
├── frontend/
│   ├── src/
│   │   ├── components/
│   │   │   ├── connection/          # 连接列表/对话框
│   │   │   ├── object-tree/         # 对象树浏览器
│   │   │   ├── object-designer/     # 表设计器
│   │   │   ├── sql-editor/          # SQL 编辑器
│   │   │   ├── data-table/          # 结果网格
│   │   │   ├── layout/              # 应用外壳
│   │   │   └── ui/                  # shadcn/ui 原语
│   │   ├── lib/api/                 # Wails 绑定封装
│   │   └── stores/                  # Zustand 状态管理
│   └── wailsjs/                     # 自动生成（勿手动编辑）
└── doc/
    └── DBStudio-product-requirements.md
```

## 📋 路线图

| 版本 | 目标 |
|------|------|
| **V1.0** | Navicat 风格对象树 + 表设计器 + MySQL/PG 基础 CRUD |
| **V1.1** | 索引设计器、Quick-Open Table、结构对比 |
| **V1.5** | 外键设计器、重命名/清空表、视图管理 |
| **V2.0** | ER 建模、反向建模、跨库同步、导入导出向导 |

## 📜 开源协议

MIT License
