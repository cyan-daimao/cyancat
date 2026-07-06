# AGENTS.md — cyancat / DBStudio

> 本文件面向 AI 编码助手。阅读前假设你对本项目一无所知。内容基于当前仓库实际结构、代码与配置整理，请勿凭经验泛化。

---

## 1. 项目概述

**cyancat**（对外产品名 **DBStudio**）是一款面向开发者、数据工程师和轻量 DBA 的跨平台数据库桌面客户端，采用 Navicat 风格的对象树 + 表设计器 + SQL 编辑器交互模型。

核心能力：

- 多数据源连接管理（MySQL / MariaDB / PostgreSQL / SQLite / StarRocks）。
- 对象树浏览器：逐层懒加载 database/schema/table/字段/索引/外键。
- 可视化表设计器：字段、索引、外键网格编辑，所有结构变更先预览 DDL 再执行。
- SQL 编辑器：基于 Monaco Editor，支持执行、格式化、注释切换。
- 结果网格：虚拟滚动、列宽拖拽、分页，导出 TSV / INSERT SQL / CSV。
- 安全确认：高风险 DDL（DROP TABLE / DROP DATABASE）强制二次确认。
- 凭据加密：AES-GCM 加密存储连接密码，生产环境推荐配置主密钥。
- 跨平台桌面应用：macOS / Windows / Linux。

源码仓库：[https://github.com/cyan-daimao/cyancat](https://github.com/cyan-daimao/cyancat)  
许可证：MIT

---

## 2. 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | Wails v2（Go + 原生 WebView） |
| 后端 | Go 1.25+，分层架构（adapter / application / domain / infra） |
| 本地存储 | SQLite（GORM），路径 `~/.cyancat/cyancat.db` |
| 数据库驱动 | go-sql-driver/mysql、jackc/pgx/v5、mattn/go-sqlite3 |
| 日志 | rs/zerolog |
| 前端 | React 18 + TypeScript + Vite |
| 样式 | Tailwind CSS + shadcn/ui（Radix UI 原语） |
| 状态管理 | Zustand |
| 编辑器 | Monaco Editor（`@monaco-editor/react`） |
| 表格 | TanStack Table + TanStack Virtual |

---

## 3. 架构与代码组织

### 3.1 整体目录结构

```
cyancat/
├── main.go                    # 入口 + 依赖注入（DI）组装
├── wails.json                 # Wails 应用配置
├── go.mod / go.sum            # Go 依赖
├── internal/                  # Go 后端源代码
│   ├── adapter/               # Wails API 绑定、DTO、转换函数
│   ├── application/           # 应用服务（连接 / 查询 / Schema）
│   ├── domain/                # 领域模型与仓库接口
│   └── infra/                 # 基础设施实现
├── frontend/                  # 前端源码
│   ├── src/
│   │   ├── components/        # 按功能组织的 React 组件
│   │   ├── lib/api/           # Wails 绑定封装
│   │   ├── stores/            # Zustand 状态管理
│   │   └── ...
│   └── wailsjs/               # Wails 自动生成的 JS/TS 绑定（禁止手动编辑）
├── scripts/build-all.sh       # 一键交叉编译脚本
├── doc/                       # 产品需求与构建文档
│   ├── BUILD.md               # macOS/Windows 打包指南
│   └── DBStudio-product-requirements.md
└── README.md
```

### 3.2 DDBD 四层架构

```
adapter → application → domain → infra
```

| 层 | 路径 | 职责 |
|----|------|------|
| **adapter** | `internal/adapter/` | Wails 绑定入口、HTTP API 结构体（实为命名约定）、DTO 与转换函数 |
| **application** | `internal/application/<biz>service/` | 服务接口与实现、Cmd/Query/BO 对象、编排业务逻辑 |
| **domain** | `internal/domain/<biz>/` | 充血模型实体、Repository 接口 |
| **infra** | `internal/infra/` | GORM 仓库实现、数据库驱动抽象、会话管理、加密、日志、事件总线 |

当前三个垂直切片：

- `connectionservice`：连接 CRUD、测试连接、打开/关闭会话。
- `queryservice`：SQL 执行、流式结果、查询历史。
- `schemaservice`：Schema 浏览、DDL 生成与执行、表设计器支撑。

### 3.3 驱动抽象（核心基础设施）

`internal/infra/driver/driver.go` 定义数据库驱动抽象：

- `Driver` / `Conn`：建立连接、执行 SQL、流式查询。
- `Dialect`：方言差异（标识符引号、参数占位符、默认 LIMIT）。
- `Inspector`：元数据查询（database / schema / table / 字段 / 索引 / 外键）。
- `DDLGenerator`：方言专属 DDL 生成。
- `RowStream`：大结果集流式游标。

已注册驱动：

- `internal/infra/driver/mysql`
- `internal/infra/driver/postgres`
- `internal/infra/driver/sqlite`
- `internal/infra/driver/starrocks`（MySQL 协议兼容）

驱动注册在 `main.go` 中完成：

```go
driver.Register(mysqldriver.New())
driver.Register(postgresdriver.New())
driver.Register(sqlitedriver.New())
driver.Register(starrocksdriver.New())
```

### 3.4 数据转换链

每层边界使用显式 `ToXxx` 转换函数，**不使用反射**：

```
读取:  DO → Domain → BO → DTO
写入:  Request → Cmd → Domain → Repository → DO
```

转换函数位置：

- `adapter/dto/*_dto.go`：Request↔Cmd，BO→DTO。
- `application/<biz>service/*_convert.go`：Cmd→Domain，Domain→BO。
- `infra/db/<biz>repo/*_convert.go`：DO↔Domain。

### 3.5 关键运行时组件

- **`infra/session/manager.go`**：运行时连接池，维护 `connectionID → driver.Conn` 的长连接。注意：`session.Manager` 属于基础设施层，不是领域层。
- **`infra/eventbus/bus.go`**：封装 Wails `runtime.EventsEmit`，用于后端向前端推送流式查询事件（`query:rows`、`query:done`、`query:error`、`connection:state`）。`Init(ctx)` 必须在 `OnStartup` 中调用，因为 `ctx` 在启动前不可用。
- **`infra/crypto/aes.go`**：AES-GCM 加解密，用于连接密码存储。
- **`infra/keychain/keychain.go`**：OS Keychain 封装（V1.0 当前为 AES 兜底实现）。
- **`infra/logger/logger.go`**：zerolog 日志，应用退出时显式 `Close()` 以保证 Windows GUI 子系统下日志落盘。

### 3.6 前端组织

- `frontend/src/lib/api/*.ts`：封装 `wailsjs/go/http/*` 绑定，统一解包 `{code, message, data}`，非 200 时抛出错误。
- `frontend/src/stores/*.ts`：Zustand 状态管理（`connection`、`query`、`schema`、`designer`、`sql-hints`）。
- `frontend/src/components/connection/`：连接列表与对话框。
- `frontend/src/components/object-tree/`：对象树浏览器与右键菜单。
- `frontend/src/components/object-designer/`：表设计器、DDL 预览、风险确认。
- `frontend/src/components/sql-editor/`：SQL 编辑器工作区。
- `frontend/src/components/data-table/`：结果网格。
- `frontend/src/components/ui/`：shadcn/ui 基础组件。
- 路径别名 `@` 指向 `frontend/src`。

---

## 4. 构建与运行命令

### 4.1 前置依赖

| 依赖 | 版本/说明 |
|------|----------|
| Go | 1.25+ |
| Node.js | 现代 LTS |
| Wails CLI | v2.12.0（`go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0`） |
| macOS | WebKit（系统内置）、Xcode Command Line Tools |
| Windows（交叉编译） | mingw-w64（`brew install mingw-w64`） |
| Linux（交叉编译） | musl-cross（`brew install FiloSottile/musl-cross/musl-cross`） |

自检命令：

```bash
wails doctor
which x86_64-w64-mingw32-gcc
```

### 4.2 开发模式

```bash
# 启动后端热重载 + 前端热更新
wails dev
```

### 4.3 构建

```bash
# 当前平台
wails build

# 指定平台
wails build -platform darwin/universal    # macOS 通用二进制（推荐分发）
wails build -platform darwin/arm64        # Apple Silicon
wails build -platform darwin/amd64        # Intel
wails build -platform windows/amd64       # Windows
wails build -platform linux/amd64         # Linux
```

Windows 交叉编译需要显式指定 C 工具链：

```bash
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
  wails build -platform windows/amd64 -clean
```

产物输出到 `build/bin/`。

### 4.4 一键脚本

```bash
# 默认构建 darwin/arm64 + darwin/amd64 + windows/amd64，归档到 build/dist/
./scripts/build-all.sh

# 指定平台
./scripts/build-all.sh darwin/universal
./scripts/build-all.sh windows/amd64 linux/amd64
```

### 4.5 前端独立命令

```bash
cd frontend
npm install
npm run dev       # 独立 Vite 开发服务器
npm run build     # tsc && vite build → frontend/dist
npm run preview   # 预览生产构建
```

### 4.6 Go 独立命令

```bash
go build ./...           # 编译检查
go vet ./...             # 静态检查
go test ./...            # 运行测试
go mod tidy              # 整理依赖
wails generate module    # 修改 Go API 后重新生成 frontend/wailsjs/ 绑定
```

> 注意：`wails generate module`、`wails build` 会生成 `frontend/wailsjs/`、`frontend/dist/`、`build/dist/` 等目录，这些已被 `.gitignore` 忽略，请勿手动提交。

---

## 5. 开发约定

### 5.1 分层约定

1. **无 Gin / 无 HTTP 路由。** `adapter/http` 只是命名约定，方法通过 Wails 反射绑定给前端，不是真正的 HTTP 路由。
2. **仅 `connection` 拥有独立的 `domain` 包。** `query` 和 `schema` 是驱动执行型关注点，没有持久化聚合；查询历史是基础设施读模型，`HistoryRepository` 接口声明在 `application/queryservice`，不要强行创建空的 domain 包。
3. **`session.Manager` 属于 infra 层**，不是 domain 层。
4. **`connectionrepo.Update` 使用 `map[string]interface{}` Updates 调用**，以确保零值字段（如 `ssl=false`）能被写入，不要改成结构体更新。
5. **`frontend/wailsjs/` 是 Wails 自动生成目录**，禁止手动编辑；修改 Go API 签名后运行 `wails generate module` 重新生成。
6. **转换函数返回空切片而非 nil。** `make([]*X, 0)`，因为 `nil` 序列化为 JSON `null` 会导致前端 `.map()` 调用崩溃。
7. **字段命名一致性。** Go 结构体使用 `durationMs` 等 json tag，前端 TypeScript 类型必须与 `wailsjs/go/models.ts` 中的生成名称完全一致。

### 5.2 代码风格

- Go：使用标准 Go 格式，`go fmt` / `gofumpt`。
- TypeScript：使用项目内 `tsconfig.json` 与 Vite 配置。
- 注释：公共包、接口、结构体字段需写中文或英文 Go doc 注释；本项目源码注释以中文为主。
- 错误处理：Go 侧返回错误，adapter 层统一包装为 `infra/api.Response[T]`；前端在 `lib/api/*.ts` 中解包并抛出。

### 5.3 大整数精度

查询结果从后端返回前端时，所有单元格统一格式化为字符串（`string | null`），避免 JavaScript `Number` 类型 53 位整数精度限制导致 `int64` / `bigint` 截断。列定义仍保留原始 `databaseType`，前端据此做数值/布尔/字符串格式化与 SQL 字面量生成。

### 5.4 DDL 执行原则

- 所有结构变更（建库、建表、改表、删表、删库）必须先展示预览 DDL，由用户确认后再执行。
- 高风险操作（`DROP TABLE`、`DROP DATABASE`）必须弹出二次确认对话框。

---

## 6. 测试说明

### 6.1 Go 测试

当前测试文件分布：

```
internal/adapter/dto/query_dto_test.go
internal/domain/connection/connection_sqlite_test.go
internal/infra/driver/value_test.go
internal/infra/driver/driver_test.go
internal/infra/driver/sqlite/driver_test.go
internal/infra/driver/postgres/driver_test.go
internal/infra/driver/starrocks/driver_test.go
```

运行全部测试：

```bash
go test ./...
```

部分测试可能需要实际数据库连接，未配置时会跳过或失败，请查看具体测试文件中的环境变量要求。

### 6.2 前端测试

当前前端没有单元测试（`*.test.ts` / `*.test.tsx` 数量为 0）。如需新增，可在 `frontend/` 内按项目现有目录结构放置测试文件。

### 6.3 验证清单

提交前建议执行：

```bash
go build ./...
go vet ./...
go test ./...
cd frontend && npm run build
```

---

## 7. 安全注意事项

### 7.1 主密钥配置

连接密码使用 AES-GCM 加密存储。加密主密钥获取优先级：

1. 环境变量 `CYANCAT_MASTER_KEY`：64 位十六进制字符串（32 字节）。
2. 文件 `~/.cyancat/master.key`：32 字节原始密钥文件。
3. 开发兜底：硬编码默认密钥（**仅开发使用**，会在日志中输出警告）。

生产环境必须配置 `CYANCAT_MASTER_KEY` 或 `~/.cyancat/master.key`，严禁依赖默认密钥。

### 7.2 本地数据

- 本地 SQLite 数据库：`~/.cyancat/cyancat.db`，自动创建。
- 主密钥文件 `~/.cyancat/master.key` 属于敏感文件，已加入 `.gitignore`，请勿提交到版本控制。

### 7.3 密码处理

- `domain/connection.Connection.Password` 在领域对象中临时持有明文；持久化时由 `infra/db/connectionrepo` 调用 `crypto.Encrypt(masterKey)` 加密后存入 `password_encrypted` 字段。
- 不要在前端日志、错误信息或网络响应中暴露密码。

### 7.4 DDL 安全

- 后端不对 DDL 做隐式过滤或重写，所有执行均基于用户确认后的 SQL。
- 高风险操作强制二次确认由前端触发，但后端仍需把权限控制在最小必要范围。

---

## 8. 关键文件速查

| 文件 | 说明 |
|------|------|
| `main.go` | 应用入口、驱动注册、本地 SQLite 初始化、DI 组装、Wails 启动 |
| `wails.json` | Wails 应用元数据、构建命令、产物名称 |
| `go.mod` | Go 依赖清单 |
| `frontend/package.json` | 前端依赖与脚本 |
| `frontend/vite.config.ts` | Vite 配置（含路径别名 `@`） |
| `frontend/tailwind.config.js` | Tailwind CSS 配置 |
| `scripts/build-all.sh` | 多平台交叉编译脚本 |
| `doc/BUILD.md` | macOS + Windows 打包详细指南 |
| `doc/DBStudio-product-requirements.md` | V1.x 产品需求与路线图 |

---

## 9. 常见操作速查

| 目标 | 命令 |
|------|------|
| 开发运行 | `wails dev` |
| 当前平台构建 | `wails build` |
| 生成 Wails 绑定 | `wails generate module` |
| 运行测试 | `go test ./...` |
| 静态检查 | `go vet ./...` |
| 交叉编译三平台 | `./scripts/build-all.sh` |
| 仅构建 Windows | `./scripts/build-all.sh windows/amd64` |
| 验证 universal 双架构 | `lipo -info build/bin/cyancat.app/Contents/MacOS/cyancat` |

---

> 最后更新：基于 2026-07-06 仓库状态整理。若项目结构发生重大变化，请同步更新本文件。
