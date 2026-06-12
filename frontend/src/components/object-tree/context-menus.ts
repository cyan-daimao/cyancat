import type { TreeNode } from '@/stores/schema';

/** 菜单上下文：传给 disabled 判断器 */
export interface MenuContext {
  node: TreeNode;
  isConnOpen: boolean;
}

/** 单个菜单项定义 */
export interface MenuItemDef {
  /** 菜单文案 */
  label?: string;
  /** 语义 action 名，由 ObjectTreeContextMenu 转发 */
  action?: string;
  /** lucide-react 图标名 */
  icon?: string;
  /** 危险操作（红色样式） */
  danger?: boolean;
  /** 动态禁用 */
  disabled?: (ctx: MenuContext) => boolean;
  /** 是否为分隔线 */
  separator?: boolean;
  /** 子菜单项（用于"生成 SQL >"等二级菜单） */
  children?: MenuItemDef[];
}

const SEP: MenuItemDef = { separator: true };

const needsOpenConn = (ctx: MenuContext) => !ctx.isConnOpen;

/** 各节点类型的右键菜单配置 */
export const contextMenuConfig: Record<string, MenuItemDef[]> = {
  // ---- 连接节点 ----
  connection: [
    { label: '打开连接', action: 'open-connection', icon: 'Plug', disabled: (c) => c.isConnOpen },
    { label: '关闭连接', action: 'close-connection', icon: 'PlugZap', disabled: (c) => !c.isConnOpen },
    SEP,
    { label: '新建数据库', action: 'new-database', icon: 'DatabaseZap', disabled: needsOpenConn },
    { label: '新建查询', action: 'new-query', icon: 'FileCode', disabled: needsOpenConn },
    SEP,
    { label: '刷新', action: 'refresh', icon: 'RefreshCw', disabled: needsOpenConn },
    { label: '连接属性', action: 'connection-properties', icon: 'Settings' },
    SEP,
    { label: '复制连接名称', action: 'copy-name', icon: 'Copy' },
  ],

  // ---- 数据库节点 ----
  database: [
    { label: '设为当前数据库', action: 'set-current-database', icon: 'Star' },
    { label: '新建表', action: 'new-table', icon: 'TableProperties' },
    { label: '新建查询', action: 'new-query', icon: 'FileCode' },
    SEP,
    { label: '刷新', action: 'refresh', icon: 'RefreshCw' },
    { label: '查看数据库属性', action: 'database-properties', icon: 'Info' },
    SEP,
    { label: '复制数据库名', action: 'copy-name', icon: 'Copy' },
    { label: '复制限定名称', action: 'copy-full-name', icon: 'Copy' },
    SEP,
    { label: '删除数据库', action: 'drop-database', icon: 'Trash2', danger: true },
  ],

  // ---- Schema 节点（PostgreSQL）----
  schema: [
    { label: '新建表', action: 'new-table', icon: 'TableProperties' },
    { label: '新建查询', action: 'new-query', icon: 'FileCode' },
    SEP,
    { label: '刷新', action: 'refresh', icon: 'RefreshCw' },
    SEP,
    { label: '复制 Schema 名', action: 'copy-name', icon: 'Copy' },
    { label: '复制限定名称', action: 'copy-full-name', icon: 'Copy' },
  ],

  // ---- 表节点 ----
  table: [
    { label: '打开表', action: 'open-table', icon: 'Table2' },
    { label: '设计表', action: 'design-table', icon: 'PencilRuler' },
    { label: '查看 DDL', action: 'view-ddl', icon: 'FileText' },
    { label: '复制 DDL', action: 'copy-ddl', icon: 'Copy' },
    SEP,
    {
      label: '生成 SQL',
      icon: 'Code2',
      children: [
        { label: 'SELECT', action: 'generate-sql-select' },
      ],
    },
    SEP,
    { label: '复制表名', action: 'copy-name', icon: 'Copy' },
    { label: '复制限定表名', action: 'copy-full-name', icon: 'Copy' },
    SEP,
    { label: '刷新', action: 'refresh', icon: 'RefreshCw' },
    SEP,
    { label: '删除表', action: 'drop-table', icon: 'Trash2', danger: true },
  ],

  // ---- 字段节点 ----
  column: [
    { label: '编辑字段', action: 'edit-column', icon: 'Pencil' },
    { label: '新增字段', action: 'add-column', icon: 'Plus' },
    { label: '插入字段', action: 'insert-column', icon: 'PlusSquare' },
    SEP,
    { label: '删除字段', action: 'drop-column', icon: 'Trash2', danger: true },
    { label: '重命名字段', action: 'rename-column', icon: 'Edit' },
    SEP,
    {
      label: '设置主键',
      action: 'set-primary-key',
      icon: 'Key',
      disabled: (c) => !!c.node.isPrimary,
    },
    {
      label: '取消主键',
      action: 'drop-primary-key',
      icon: 'KeyRound',
      disabled: (c) => !c.node.isPrimary,
    },
    SEP,
    { label: '复制字段名', action: 'copy-name', icon: 'Copy' },
    { label: '复制带引号字段名', action: 'copy-quoted-name', icon: 'Copy' },
  ],

  // ---- 视图节点 ----
  view: [
    { label: '打开视图', action: 'open-view', icon: 'Eye' },
    SEP,
    { label: '查看定义', action: 'view-ddl', icon: 'FileText' },
    { label: '复制定义', action: 'copy-ddl', icon: 'Copy' },
    SEP,
    {
      label: '生成 SQL',
      icon: 'Code2',
      children: [{ label: 'SELECT', action: 'generate-sql-select' }],
    },
    SEP,
    { label: '复制视图名', action: 'copy-name', icon: 'Copy' },
    { label: '复制限定视图名', action: 'copy-full-name', icon: 'Copy' },
  ],

  // ---- 文件夹节点 ----
  'columns-folder': [
    { label: '新增字段', action: 'add-column', icon: 'Plus' },
    SEP,
    { label: '刷新', action: 'refresh-table', icon: 'RefreshCw' },
  ],

  'indexes-folder': [
    { label: '新建索引', action: 'new-index', icon: 'Plus' },
    SEP,
    { label: '刷新', action: 'refresh-table', icon: 'RefreshCw' },
  ],

  'foreign-keys-folder': [
    { label: '新建外键', action: 'new-foreign-key', icon: 'Plus' },
    SEP,
    { label: '刷新', action: 'refresh-table', icon: 'RefreshCw' },
  ],

  // ---- 索引节点 ----
  index: [
    { label: '编辑索引', action: 'edit-index', icon: 'Pencil' },
    { label: '删除索引', action: 'drop-index', icon: 'Trash2', danger: true },
    SEP,
    { label: '复制索引名', action: 'copy-name', icon: 'Copy' },
  ],

  // ---- 外键节点 ----
  foreignKey: [
    { label: '编辑外键', action: 'edit-foreign-key', icon: 'Pencil' },
    { label: '删除外键', action: 'drop-foreign-key', icon: 'Trash2', danger: true },
    SEP,
    { label: '复制外键名', action: 'copy-name', icon: 'Copy' },
  ],
};

/** 根据节点类型获取菜单项配置 */
export function getMenuItemsForNode(node: TreeNode): MenuItemDef[] {
  return contextMenuConfig[node.type] || [];
}
