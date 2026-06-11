# DBStudio 产品需求文档

## 1. 背景

DBStudio 当前具备连接管理、对象树、SQL 编辑器、查询执行和结果集展示能力。下一阶段产品目标不是只补几个右键菜单，而是对标 Navicat 的数据库对象管理体验：用户围绕左侧对象树，通过右键菜单进入新建数据库、新建表、设计表、编辑字段、查看 DDL、打开表数据、复制对象名等工作流。

Navicat 的典型产品心智是：

1. Navigation Pane / 对象树是数据库对象入口。
2. 右键菜单承载对象管理动作。
3. Table Designer 是创建和编辑表结构的核心界面。
4. 表结构包括 fields、indexes、foreign keys 等对象。
5. 表数据可以通过 Open Table / Quick Open 方式打开。
6. 表 DDL 可以在对象信息或 DDL 面板中查看。

因此，DBStudio 的新需求应从“右键快捷菜单”升级为“Navicat 式对象树 + 可视化对象设计器”。

参考资料：

1. [Navicat 17 Windows PDF Manual](https://www.navicat.com/manual/pdf_manual/en/navicat_17/win_manual/navicat_en.pdf)
2. [Navicat Help - How to view table structure](https://help.navicat.com/hc/en-us/articles/217885468-How-to-view-table-structure-of-a-table)
3. [Navicat Help - Design tables without Table Designer](https://help.navicat.com/hc/en-us/articles/218299827-How-do-we-design-tables-without-using-table-designer)
4. [Navicat Help - Foreign key creation notes](https://help.navicat.com/hc/en-us/articles/217791848-Why-I-cannot-successfully-create-the-foreign-keys)

## 2. 产品定位

DBStudio 是面向开发者、数据开发和轻量 DBA 场景的数据库桌面客户端。产品形态对标 Navicat 的对象树、右键菜单、表设计器、DDL 查看、数据表打开体验，但实现策略更聚焦：

1. 先支持 MySQL/MariaDB 的高频对象管理。
2. PostgreSQL 先按 database/schema/table 模型预留。
3. 所有结构变更必须先预览 SQL。
4. 高风险 DDL 必须二次确认。
5. SQL 编辑器仍作为高级用户兜底入口。

## 3. 产品目标

1. 把对象树右键菜单改造成 Navicat 式对象操作入口。
2. 支持新建数据库、编辑数据库属性、删除数据库的规划入口。
3. 支持新建表、设计表、编辑字段、索引、外键的可视化入口。
4. 支持 Open Table / Quick Open Table 式的数据打开体验。
5. 支持查看 DDL、复制 DDL、在 SQL 编辑器打开 DDL。
6. 支持字段级右键操作：新增字段、插入字段、删除字段、设置主键、重命名字段、复制字段名。
7. 支持结果集右键复制、导出、查看完整内容。
8. 建立 DDL Preview / Execute DDL 的后端接口模型。
9. 建立对象设计器的前端组件模型。

## 4. 不做什么

V1.0 不做以下能力：

1. ER 图建模。
2. 跨库结构同步和结构对比。
3. 完整数据编辑提交。
4. 存储过程、触发器、事件、函数设计器。
5. 分区表、表空间、权限、角色管理。
6. 自动迁移脚本版本管理。
7. 多人协作和审批流。

## 5. 用户角色

### 5.1 后端开发

快速创建测试库、业务表，新增字段，修改字段类型，设置主键和自增，查看 DDL，复制表名/字段名。

### 5.2 数据开发

创建数据落地表，调整字段注释，检查字段类型，复制 DDL 到其他环境执行，快速预览数据。

### 5.3 DBA / 运维

刷新对象树，查看连接状态，断开连接，查看 DDL，评估结构变更风险，避免误删误改。

### 5.4 初级用户

不熟悉完整 DDL 语法，需要通过表单和表格安全完成数据库、表、字段维护。

## 6. Navicat 对标功能范围

本章节是本次文档重写的核心。DBStudio 的对象树右键菜单不再只是普通快捷操作，而是按 Navicat 的 Navigation Pane / Table Designer 心智组织。

### 6.1 连接节点右键菜单

连接节点承载连接生命周期、数据库创建和连接级工具入口。

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 打开连接 | Open Connection | 建立连接并加载数据库列表 | P0 |
| 关闭连接 | Close Connection | 关闭连接并清理 session/object tree cache | P0 |
| 刷新 | Refresh | 重新加载数据库/schema 列表 | P0 |
| 新建数据库 | New Database / New Schema | 打开新建数据库弹窗 | P0 |
| 新建查询 | New Query | 打开/聚焦 SQL 编辑器，并绑定当前连接 | P0 |
| 连接属性 | Connection Properties | 打开连接编辑弹窗 | P1 |
| 复制连接名称 | Copy Name | 复制连接显示名 | P1 |
| 删除连接 | Delete Connection | 二次确认后删除连接配置 | P2 |

### 6.2 数据库 / Schema 节点右键菜单

数据库节点承载库级对象管理。MySQL 下是 database，PostgreSQL 下可映射为 database/schema。

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 设为当前数据库 | Select / Use Database | 更新 `selectedNode`，右侧执行使用该 database/schema | P0 |
| 新建表 | New Table | 打开 Table Designer 新建模式 | P0 |
| 新建查询 | New Query | 聚焦 SQL 编辑器，绑定当前 database/schema | P0 |
| 刷新 | Refresh | 重新加载该库下表、视图等对象 | P0 |
| 查看数据库属性 | Database / Schema Properties | 展示字符集、排序规则等信息 | P1 |
| 编辑数据库属性 | Edit Database / Edit Schema | 打开数据库属性编辑弹窗 | P1 |
| 复制数据库名 | Copy Name | 复制数据库名 | P1 |
| 复制限定名称 | Copy Full Name | 复制带连接/database/schema 语义的限定名 | P1 |
| 删除数据库 | Delete / Drop Database | 高风险二次确认，V1.1 后支持 | P2 |

### 6.3 表节点右键菜单

表节点是最重要的右键入口。Navicat 中表节点不仅能打开数据，还能进入设计器、查看 DDL、清空/截断、复制、重命名等。

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 打开表 | Open Table | 查询并展示前 N 行数据，默认 1000 | P0 |
| 快速打开表 | Open Table (Quick) | 查询前 N 行，但延迟加载大字段/BLOB，V1.1 支持 | P1 |
| 设计表 | Design Table | 打开 Table Designer 编辑模式 | P0 |
| 新建查询 | New Query | 在当前表上下文打开 SQL 编辑器 | P0 |
| 查看 DDL | Object Information / DDL | 打开 DDL 只读预览面板 | P0 |
| 复制 DDL | Copy DDL | 复制当前表 DDL | P0 |
| 生成 SQL > SELECT | Generate SQL | 插入 SELECT 模板到 SQL 编辑器 | P0 |
| 生成 SQL > INSERT | Generate SQL | 插入 INSERT 模板 | P1 |
| 生成 SQL > UPDATE | Generate SQL | 插入 UPDATE 模板 | P1 |
| 生成 SQL > DELETE | Generate SQL | 插入 DELETE 模板 | P1 |
| 复制表名 | Copy Name | 复制表名 | P0 |
| 复制限定表名 | Copy Full Name | 复制 `database.table` 或 `schema.table` | P0 |
| 重命名表 | Rename | 打开重命名弹窗，预览 RENAME/ALTER SQL | P2 |
| 清空表 | Empty Table | 删除数据但不重置自增，V1.5 后支持 | P3 |
| 截断表 | Truncate Table | 清空并重置自增，高风险，V1.5 后支持 | P3 |
| 删除表 | Delete / Drop Table | 高风险二次确认，V1.5 后支持 | P3 |
| 刷新 | Refresh | 重新加载表结构和子节点 | P0 |

### 6.4 表子对象节点

Navicat 的表结构不只是字段，表下通常包含 fields、indexes、foreign keys 等元素。DBStudio 对象树应逐步从“表 -> 字段”升级为“表 -> Columns / Indexes / Foreign Keys”。

#### Columns 文件夹节点

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 新增字段 | Add Field | 打开 Table Designer 并在末尾新增字段行 | P0 |
| 插入字段 | Insert Field | 打开 Table Designer 并在选中位置插入字段 | P1 |
| 刷新 | Refresh | 重新加载字段列表 | P0 |

#### 字段节点

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 编辑字段 | Design Table / Edit Field | 打开 Table Designer 并聚焦该字段 | P0 |
| 新增字段 | Add Field | 在字段末尾新增字段 | P0 |
| 插入字段 | Insert Field | 在当前字段上方插入字段 | P1 |
| 删除字段 | Delete Field | 标记删除，预览 DROP COLUMN，高风险确认 | P1 |
| 重命名字段 | Rename Field | 预览 CHANGE/RENAME COLUMN SQL | P1 |
| 设置主键 | Primary Key | 切换字段 Primary Key 状态 | P0 |
| 取消主键 | Drop Primary Key | 预览修改主键 SQL | P1 |
| 复制字段名 | Copy Name | 复制字段名 | P0 |
| 复制带引号字段名 | Copy Quoted Name | 按方言复制 quoted identifier | P0 |
| 添加到编辑器 | Add to Query Editor | 在 SQL 编辑器光标处插入字段名 | P1 |

#### Indexes 文件夹节点

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 新建索引 | New Index | 打开 Table Designer 的 Indexes tab | P1 |
| 刷新 | Refresh | 重新加载索引列表 | P1 |

#### 索引节点

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 编辑索引 | Edit Index | 打开 Table Designer 并聚焦该索引 | P1 |
| 删除索引 | Delete Index | 预览 DROP INDEX，高风险确认 | P2 |
| 复制索引名 | Copy Name | 复制索引名 | P1 |

#### Foreign Keys 文件夹节点

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 新建外键 | New Foreign Key | 打开 Table Designer 的 Foreign Keys tab | P2 |
| 刷新 | Refresh | 重新加载外键列表 | P1 |

#### 外键节点

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 编辑外键 | Edit Foreign Key | 打开外键编辑器 | P2 |
| 删除外键 | Delete Foreign Key | 预览 DROP FOREIGN KEY，高风险确认 | P2 |
| 复制外键名 | Copy Name | 复制外键名 | P1 |

### 6.5 视图节点右键菜单

| 菜单项 | Navicat 对标 | DBStudio 行为 | 优先级 |
| --- | --- | --- | --- |
| 打开视图 | Open View | 查询并展示前 N 行 | P0 |
| 设计视图 | Design View | 打开视图 SQL 定义编辑器，V1.2 后支持 | P2 |
| 查看定义 | View Definition / DDL | 展示 view definition | P1 |
| 复制定义 | Copy DDL | 复制视图定义 SQL | P1 |
| 生成 SQL > SELECT | Generate SQL | 插入 SELECT 模板 | P0 |
| 复制视图名 | Copy Name | 复制视图名 | P0 |
| 复制限定视图名 | Copy Full Name | 复制限定名称 | P0 |
| 删除视图 | Drop View | 高风险二次确认，V1.5 后支持 | P3 |

### 6.6 空白区域右键菜单

对象树空白区域也需要有 Navicat 式全局入口，避免顶部按钮消失后缺少对象创建入口。

| 菜单项 | 行为 | 优先级 |
| --- | --- | --- |
| 新建连接 | 打开连接创建弹窗 | P1 |
| 刷新全部 | 重新加载连接和已打开对象树 | P1 |
| 新建查询 | 如果已有选中连接，则打开 SQL 编辑器 | P1 |

## 7. 可视化新建数据库

### 7.1 入口

1. 连接节点右键：`新建数据库`。
2. 对象树空白处右键：如果已选中连接，可显示 `新建数据库`。
3. 快捷入口后续可放在 Command Palette，不再依赖顶部按钮。

### 7.2 表单字段

| 字段 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| 数据库名 | Input | 是 | 空 | MySQL database name |
| 字符集 | Select | 否 | utf8mb4 | MySQL 优先支持 |
| 排序规则 | Select | 否 | utf8mb4_general_ci | 跟随字符集联动 |
| 创建后自动选中 | Checkbox | 否 | true | 创建成功后设置为当前 selectedNode |

### 7.3 交互

1. 用户输入数据库名。
2. 系统本地校验名称不能为空、不能包含非法字符。
3. 用户点击“预览 SQL”。
4. 系统展示 DDL 预览。
5. 用户确认执行。
6. 执行成功后刷新连接节点，并选中新数据库。

### 7.4 SQL 示例

```sql
CREATE DATABASE `demo_db`
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_general_ci;
```

PostgreSQL V1.0 先支持 schema：

```sql
CREATE SCHEMA "demo_schema";
```

## 8. Table Designer

Table Designer 是本产品的核心功能。它不只是一个弹窗，而是 DBStudio 的“表结构工作台”。

### 8.1 打开方式

| 模式 | 入口 | 行为 |
| --- | --- | --- |
| 新建模式 | 数据库右键 `新建表` | 空表结构，保存生成 CREATE TABLE |
| 编辑模式 | 表右键 `设计表` | 加载现有表结构，保存生成 ALTER TABLE |
| 字段聚焦模式 | 字段右键 `编辑字段` | 打开设计器并定位到字段行 |
| 索引聚焦模式 | 索引右键 `编辑索引` | 打开设计器并定位到索引 tab |
| 外键聚焦模式 | 外键右键 `编辑外键` | 打开设计器并定位到外键 tab |
| 查看模式 | 表右键 `查看 DDL` | 只读展示 DDL |

### 8.2 顶部表属性

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| 表名 | Input | 是 | 新建模式可编辑，编辑模式默认不可直接改名 |
| 表注释 | Input | 否 | MySQL COMMENT |
| 存储引擎 | Select | 否 | MySQL 默认 InnoDB |
| 字符集 | Select | 否 | 默认继承数据库 |
| 排序规则 | Select | 否 | 默认继承数据库 |

### 8.3 Tab 设计

| Tab | V1.0 | V1.1 | 说明 |
| --- | --- | --- | --- |
| 字段 Fields | 是 | 是 | 字段增删改、主键、自增、默认值、注释 |
| 索引 Indexes | 基础 | 是 | V1.0 支持主键/唯一索引基础展示和生成 |
| 外键 Foreign Keys | 否 | 是 | V1.1 后支持外键创建和编辑 |
| 选项 Options | 基础 | 是 | 引擎、字符集、排序规则 |
| SQL 预览 SQL Preview | 是 | 是 | 展示 CREATE/ALTER SQL |

## 9. 字段设计器

字段设计器采用可编辑表格形态，对标 Navicat Table Designer 的 Fields tab。

### 9.1 字段列定义

| 列 | 类型 | 必填 | 默认值 | 说明 |
| --- | --- | --- | --- | --- |
| 字段名 | Input | 是 | 空 | 不允许重复 |
| 类型 | Combobox | 是 | varchar | 常见类型可搜索 |
| 长度 | Input | 否 | 类型默认 | 如 varchar(255), decimal(10,2) |
| 小数位 | Input | 否 | 空 | decimal/float 适用 |
| Not Null | Checkbox | 否 | false | 是否非空 |
| Primary Key | Checkbox | 否 | false | 是否主键 |
| Auto Increment | Checkbox | 否 | false | MySQL 整数主键适用 |
| Unsigned | Checkbox | 否 | false | MySQL 数字类型适用 |
| Default | Input/Select | 否 | 空 | 支持 NULL、空字符串、函数文本 |
| Comment | Input | 否 | 空 | 字段注释 |
| 操作 | Icon buttons | 否 | - | 上移、下移、删除、复制 |

### 9.2 字段操作

| 操作 | 行为 | 优先级 |
| --- | --- | --- |
| 新增字段 | 在字段末尾插入一行 | P0 |
| 插入字段 | 在当前字段上方插入一行 | P1 |
| 删除字段 | 标记删除，保存时生成 DROP COLUMN | P0 |
| 复制字段 | 复制当前字段定义为新行 | P1 |
| 上移/下移 | 调整字段顺序 | P1 |
| 设置主键 | 切换 Primary Key | P0 |
| 设置自增 | 仅整数类型允许 | P0 |
| 重命名字段 | 修改字段名并生成 rename/change SQL | P1 |
| 搜索字段 | 按字段名过滤/定位 | P1 |

### 9.3 校验规则

1. 字段名不能为空。
2. 同一表内字段名不能重复。
3. 主键字段自动 Not Null。
4. Auto Increment 只能用于整数类型。
5. varchar/char 必须有长度。
6. decimal 的小数位不能大于总长度。
7. 删除已有字段属于高风险操作，保存时必须明确列出。
8. 修改字段类型、长度、Not Null、默认值可能导致数据丢失，需要风险提示。

## 10. 索引设计器

V1.0 做基础展示和主键/唯一索引生成，V1.1 做完整索引编辑。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| 索引名 | Input | 必填 |
| 索引类型 | Select | PRIMARY / UNIQUE / NORMAL / FULLTEXT |
| 字段列表 | Multi select | 支持排序 |
| 注释 | Input | 可选 |

操作：

1. 新建索引。
2. 编辑索引。
3. 删除索引。
4. 从字段勾选 Primary Key 自动同步主键索引。

## 11. 外键设计器

V1.1 后支持。Navicat 的外键能力依赖字段类型、引用表、引用字段和 on delete/on update 策略。

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| 外键名 | Input | 必填 |
| 本表字段 | Select | 必须类型兼容 |
| 引用数据库/schema | Select | 默认当前库 |
| 引用表 | Select | 必填 |
| 引用字段 | Select | 必填 |
| On Delete | Select | RESTRICT / CASCADE / SET NULL / NO ACTION |
| On Update | Select | RESTRICT / CASCADE / SET NULL / NO ACTION |

校验：

1. 本表字段和引用字段类型必须兼容。
2. SET NULL 要求字段允许 NULL。
3. MySQL/MariaDB 下外键相关表建议为 InnoDB。
4. 引用字段需要有索引。

## 12. DDL 预览与执行

### 12.1 展示规则

1. 新建数据库展示 CREATE DATABASE/CREATE SCHEMA。
2. 新建表展示 CREATE TABLE。
3. 编辑表展示 ALTER TABLE。
4. 修改索引展示 ADD/DROP INDEX。
5. 修改外键展示 ADD/DROP FOREIGN KEY。
6. SQL 使用只读代码编辑器展示。
7. 支持复制 SQL。
8. 支持在 SQL 编辑器中打开。

### 12.2 保存流程

```text
用户编辑结构
  -> 前端校验
  -> 生成结构 diff
  -> 后端生成 DDL
  -> 展示 SQL 预览
  -> 用户确认执行
  -> 后端执行
  -> 刷新对象树和表结构缓存
```

### 12.3 风险分级

| 风险 | 示例 | 处理 |
| --- | --- | --- |
| 低 | 新增 nullable 字段 | 直接预览 |
| 中 | 修改字段注释、默认值 | 预览提示 |
| 高 | 修改字段类型、删除字段、Not Null 变更 | 二次确认 |
| 极高 | 删除表、清空表、截断表 | V1.0 不提供 |

### 12.4 SQL 示例

```sql
CREATE TABLE `bi_chart` (
  `id` bigint NOT NULL AUTO_INCREMENT COMMENT '主键',
  `name` varchar(128) NOT NULL COMMENT '图表名称',
  `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='图表表';
```

```sql
ALTER TABLE `bi_chart`
  ADD COLUMN `description` varchar(255) NULL COMMENT '描述',
  MODIFY COLUMN `name` varchar(256) NOT NULL COMMENT '图表名称',
  DROP COLUMN `legacy_code`;
```

## 13. 打开表与数据查看

对标 Navicat 的 Open Table / Open Table (Quick)。

| 功能 | 行为 | 优先级 |
| --- | --- | --- |
| 打开表 | 查询前 N 行，默认 1000，并显示在结果 tab | P0 |
| 快速打开表 | 延迟加载 BLOB/TEXT 大字段，V1.1 支持 | P1 |
| 表数据过滤 | 在结果区提供 WHERE 条件入口，V1.2 支持 | P2 |
| 表数据排序 | 点击列头排序，V1.2 支持 | P2 |

## 14. SQL 编辑器右键菜单

| 菜单项 | 行为 | 优先级 |
| --- | --- | --- |
| 执行选中/全部 | 复用现有执行逻辑 | P0 |
| 格式化 SQL | 对当前 SQL 做基础格式化 | P1 |
| 注释/取消注释 | 对选中行切换注释 | P1 |
| 复制 | 调用系统复制 | P0 |
| 粘贴 | 调用系统粘贴 | P0 |
| 剪切 | 调用系统剪切 | P0 |
| 全选 | 调用 Monaco 全选 | P0 |
| 清空编辑器 | 二次确认或撤销可恢复 | P2 |

## 15. 结果集右键菜单

| 菜单项 | 行为 | 优先级 |
| --- | --- | --- |
| 复制单元格 | 复制当前单元格文本 | P0 |
| 复制行 | 复制当前行，默认 TSV | P0 |
| 复制列名 | 复制当前列名 | P0 |
| 复制为 INSERT | 基于当前行生成 INSERT SQL | P1 |
| 复制为 Markdown | 复制当前结果或选中区域为 Markdown 表格 | P1 |
| 导出 CSV | 复用现有 CSV 导出 | P0 |
| 查看完整内容 | 长文本打开弹窗查看 | P1 |
| 以该值筛选 | 在编辑器插入 `WHERE column = value` 条件 | P2 |

## 16. 交互规范

### 16.1 右键菜单

1. 右键对象树节点时，该节点自动成为当前选中节点。
2. 菜单出现在鼠标位置，靠近窗口边缘时自动翻转。
3. 不可用操作置灰，不隐藏。
4. 危险操作使用红色文本或危险样式。
5. 执行中状态下，相关执行类菜单项置灰。
6. 连接未打开时，需要连接态的菜单项置灰或提示先打开连接。

### 16.2 Table Designer

1. 使用大尺寸 dialog 或独立工作区，宽度不小于 960px。
2. 字段表格固定表头。
3. 字段名和类型列始终可见。
4. 底部固定操作区：`取消`、`预览 SQL`、`保存`。
5. 有未保存改动时关闭弹窗必须确认。
6. 保存成功后 toast：`表结构已保存`。

### 16.3 字段编辑

1. 新增字段后自动聚焦字段名。
2. 类型选择支持搜索。
3. 勾选主键时自动勾选 Not Null。
4. 勾选自增时如果类型不合法，提示并阻止。
5. 删除字段只是标记删除，可撤销；保存后才执行。
6. 修改已有字段用视觉标识显示 dirty 状态。

## 17. 后端接口需求

建议新增 `DDLAPI` 或 `ObjectAPI`。

### 17.1 PreviewDDL

```ts
interface PreviewDDLRequest {
  connID: number;
  database?: string;
  schema?: string;
  action:
    | 'createDatabase'
    | 'alterDatabase'
    | 'createTable'
    | 'alterTable'
    | 'renameTable'
    | 'dropTable';
  payload: CreateDatabasePayload | TableDesignPayload;
}

interface PreviewDDLResponse {
  sql: string;
  warnings: DDLWarning[];
  riskLevel: 'low' | 'medium' | 'high' | 'critical';
}
```

### 17.2 ExecuteDDL

```ts
interface ExecuteDDLRequest {
  connID: number;
  database?: string;
  schema?: string;
  sql: string;
  confirmRiskToken?: string;
}

interface ExecuteDDLResponse {
  success: boolean;
  affectedObject?: {
    type: 'database' | 'schema' | 'table' | 'column' | 'index' | 'foreignKey';
    name: string;
  };
}
```

### 17.3 DescribeTableForDesign

```ts
interface DesignColumnDTO {
  name: string;
  type: string;
  length?: number;
  scale?: number;
  nullable: boolean;
  primaryKey: boolean;
  autoIncrement: boolean;
  unsigned: boolean;
  defaultValue?: string;
  comment?: string;
  ordinalPosition: number;
}
```

## 18. 前端模块建议

```text
frontend/src/components/context-menu/
  AppContextMenu.tsx
  objectTreeMenu.ts
  editorMenu.ts
  resultGridMenu.ts

frontend/src/components/object-designer/
  CreateDatabaseDialog.tsx
  TableDesignerDialog.tsx
  FieldGrid.tsx
  IndexGrid.tsx
  ForeignKeyGrid.tsx
  TableOptionsForm.tsx
  DDLPreviewPanel.tsx
  RiskConfirmDialog.tsx

frontend/src/stores/objectDesigner.ts
frontend/src/lib/sql/types.ts
```

建议使用 `@radix-ui/react-context-menu` 建立统一右键菜单组件。

## 19. 对象树数据结构建议

对象树应从简单 database/table/column 升级为更接近 Navicat 的对象树：

```text
Connection
  Database / Schema
    Tables
      Table
        Columns
          Column
        Indexes
          Index
        Foreign Keys
          Foreign Key
    Views
      View
```

V1.0 可以先不显示 Tables/Views 文件夹，直接展示表和视图；但 Table 子级建议尽早按 Columns/Indexes/Foreign Keys 组织。

## 20. 验收标准

### 20.1 对象树右键

1. 连接节点右键包含打开、关闭、刷新、新建数据库、新建查询、连接属性。
2. 数据库节点右键包含新建表、设为当前数据库、新建查询、刷新、查看属性。
3. 表节点右键包含打开表、设计表、查看 DDL、复制 DDL、生成 SQL、刷新。
4. 字段节点右键包含编辑字段、新增字段、插入字段、删除字段、重命名字段、设置主键。
5. 右键节点时，该节点成为当前 selectedNode。

### 20.2 新建数据库

1. 用户可从连接节点右键打开新建数据库弹窗。
2. 数据库名为空时不能保存。
3. MySQL 可选择字符集和排序规则。
4. 点击预览能看到 CREATE DATABASE SQL。
5. 执行成功后左侧对象树出现新数据库。

### 20.3 新建表

1. 用户可从数据库节点右键打开 Table Designer 新建模式。
2. 至少可添加字段名、类型、长度、Not Null、主键、自增、默认值、注释。
3. 字段名重复时不能保存。
4. 点击预览能看到 CREATE TABLE SQL。
5. 执行成功后左侧对象树出现新表。

### 20.4 编辑字段

1. 用户可从表节点右键打开设计表。
2. 用户可从字段节点右键打开设计表并定位字段。
3. 现有字段能正确回填。
4. 新增字段能生成 ADD COLUMN。
5. 修改字段能生成 MODIFY/ALTER COLUMN。
6. 删除字段能生成 DROP COLUMN，并触发高风险确认。
7. 保存成功后重新加载表结构。

### 20.5 DDL 预览

1. 所有结构变更保存前必须展示 DDL。
2. SQL 可复制。
3. SQL 可打开到 SQL 编辑器。
4. 高风险 SQL 必须二次确认。
5. 执行失败时设计器不关闭，用户可继续修改。

### 20.6 打开表

1. 表节点右键“打开表”后，结果区出现新 tab。
2. 查询使用表所在 database/schema。
3. 默认限制返回 1000 行。
4. 表名按方言正确 quote。

### 20.7 结果集

1. 右键单元格复制单元格，内容与显示值一致。
2. 右键行复制行，字段顺序与结果列顺序一致。
3. 右键导出 CSV，行为与现有导出入口一致。
4. 长文本查看不会撑破表格布局。

## 21. 版本规划

### V1.0：Navicat 式对象树与基础设计器

1. 对象树右键菜单重做。
2. MySQL 新建数据库。
3. MySQL 新建表。
4. MySQL 设计表和编辑字段。
5. 表打开和查询前 1000 行。
6. 查看 DDL / 复制 DDL。
7. 字段级新增、插入、删除、重命名、设置主键入口。
8. DDL 预览和风险确认。

### V1.1：索引、快速打开、PostgreSQL schema

1. 索引设计器。
2. 唯一索引、普通索引、复合主键。
3. Open Table (Quick)。
4. PostgreSQL schema 新建和基础表设计。
5. 保存前结构 diff 检测。

### V1.5：外键和危险操作

1. 外键设计器。
2. 表重命名。
3. 字段顺序调整落库。
4. 清空表、截断表、删除表。
5. 删除视图、删除索引、删除外键。

### V2.0：模型化设计

1. ER 图。
2. 反向生成模型。
3. 模型生成 DDL。
4. 结构对比和同步。
5. 跨库复制表。
6. 导入导出向导。

## 22. 风险与约束

1. DDL 操作不可轻易回滚，必须以 SQL 预览和确认机制降低风险。
2. 字段修改可能导致数据截断、转换失败或锁表。
3. 不同数据库 DDL 差异很大，V1.0 先从 MySQL 收敛。
4. 前端直接拼 SQL 风险较高，建议后端生成 DDL。
5. Table Designer UI 密度高，需要优先保证字段表格可读、可编辑、不会横向挤压。
6. Monaco 编辑器有自身上下文菜单，自定义菜单需要避免快捷键和焦点冲突。
7. 右键菜单操作依赖当前连接上下文，必须保证 `selectedNode` 更新及时。
8. 剪贴板能力在 Wails 和浏览器开发环境下可能存在差异，需要封装统一 Clipboard API。

## 23. 开放问题

1. V1.0 是否只支持 MySQL，PostgreSQL 是否先只支持 schema 和查看 DDL？
2. CREATE DATABASE 是否允许直接执行，还是必须先预览 SQL？
3. 字段类型列表是否先内置常用类型，还是从后端按 driver 返回？
4. 删除字段是否需要要求用户输入字段名确认？
5. 保存 ALTER TABLE 前是否需要展示预计影响，例如涉及删除字段、修改类型、Not Null 变更？
6. Table Designer 使用 dialog，还是独立工作区？
7. Open Table 默认 N 是否固定为 1000，还是跟随用户设置？
8. 是否需要在对象树中显示 Tables/Views/Columns/Indexes/Foreign Keys 文件夹层级？
