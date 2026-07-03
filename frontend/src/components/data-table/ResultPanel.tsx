import React from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { Button } from '@/components/ui/button';
import { Download, Copy, FileText, Clipboard, Hash, ChevronLeft, ChevronRight, ChevronsLeft, ChevronsRight } from 'lucide-react';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import type { QueryResultDTO } from '@/lib/api/types';
import { toast } from '@/components/ui/use-toast';
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuSeparator,
  ContextMenuTrigger,
} from '@/components/ui/context-menu';
import { useSchemaStore } from '@/stores/schema';

interface ResultPanelProps {
  result: QueryResultDTO;
  /** 可选：用于生成 INSERT SQL 时的表名；不传则尝试从全局 schema store 取 */
  tableName?: string;
}

/** 将任意值格式化为字符串（NULL/undefined 统一显示为 "NULL"） */
const formatValue = (v: string | null | undefined): string => {
  if (v === null || v === undefined) return 'NULL';
  return v;
};

/** 判断数据库类型是否为数值类型 */
const isNumericType = (databaseType: string): boolean => {
  const t = databaseType.toLowerCase();
  return /\b(int|integer|bigint|smallint|tinyint|mediumint|decimal|numeric|float|double|real|money|serial|bigserial|smallserial)\b/.test(t);
};

/** 判断数据库类型是否为布尔类型 */
const isBooleanType = (databaseType: string): boolean => {
  const t = databaseType.toLowerCase();
  return t === 'bool' || t === 'boolean';
};

/** 把值序列化成 SQL 字面量（NULL 保持 NULL，数值/布尔按列类型输出裸值，其余加单引号并转义） */
const toSqlLiteral = (v: string | null | undefined, databaseType: string): string => {
  if (v === null || v === undefined) return 'NULL';
  if (isBooleanType(databaseType)) return v.toLowerCase() === 'true' ? 'TRUE' : 'FALSE';
  if (isNumericType(databaseType)) return v;
  const s = v.replace(/'/g, "''");
  return `'${s}'`;
};

const PAGE_SIZE_OPTIONS = [50, 100, 500];

const DEFAULT_COL_WIDTH = 200;
const MIN_COL_WIDTH = 60;
const MAX_COL_WIDTH = 800;
const ROW_INDEX_WIDTH = 56;

const ResultPanel: React.FC<ResultPanelProps> = ({ result, tableName }) => {
  const parentRef = React.useRef<HTMLDivElement>(null);

  const cols = result.columns ?? [];
  const rows = result.rows ?? [];

  // 每列宽度：按列名 key 存储，未设置时回退到默认宽度
  const [colWidths, setColWidths] = React.useState<Record<string, number>>({});
  const getColWidth = (name: string) => colWidths[name] ?? DEFAULT_COL_WIDTH;

  // 当列集合变化时（例如切换查询结果），重置宽度
  const colsKey = React.useMemo(() => cols.map(c => c.name).join('|'), [cols]);
  React.useEffect(() => {
    setColWidths({});
  }, [colsKey]);

  // 拖拽列宽
  const dragState = React.useRef<{ name: string; startX: number; startWidth: number } | null>(null);

  const onResizeMouseDown = (e: React.MouseEvent, name: string) => {
    e.preventDefault();
    e.stopPropagation();
    dragState.current = {
      name,
      startX: e.clientX,
      startWidth: getColWidth(name),
    };
    const prevCursor = document.body.style.cursor;
    const prevSelect = document.body.style.userSelect;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';

    const onMove = (ev: MouseEvent) => {
      const s = dragState.current;
      if (!s) return;
      const next = Math.min(MAX_COL_WIDTH, Math.max(MIN_COL_WIDTH, s.startWidth + (ev.clientX - s.startX)));
      setColWidths(prev => ({ ...prev, [s.name]: next }));
    };
    const onUp = () => {
      dragState.current = null;
      document.body.style.cursor = prevCursor;
      document.body.style.userSelect = prevSelect;
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
    };
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  };

  const onResizeDoubleClick = (e: React.MouseEvent, name: string) => {
    e.preventDefault();
    e.stopPropagation();
    setColWidths(prev => {
      const next = { ...prev };
      delete next[name];
      return next;
    });
  };

  // 分页状态
  const [pageSize, setPageSize] = React.useState(50);
  const [currentPage, setCurrentPage] = React.useState(1);

  const totalPages = Math.max(1, Math.ceil(rows.length / pageSize));
  // 当数据变化时修正页码
  React.useEffect(() => {
    setCurrentPage(p => Math.min(p, totalPages));
  }, [totalPages]);

  const offset = (currentPage - 1) * pageSize;
  const pagedRows = rows.slice(offset, offset + pageSize);
  // 只要有数据就显示分页组件（即使总行数 ≤ 一页）
  const showPagination = rows.length > 0;

  // 从全局 store 推断表名（INSERT SQL 用），若未传 prop
  const selectedNode = useSchemaStore(s => s.selectedNode);
  const resolvedTableName =
    tableName ||
    selectedNode?.tableName ||
    'table_name';

  const virtualizer = useVirtualizer({
    count: pagedRows.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => 28,
    overscan: 20,
  });

  const exportCSV = () => {
    if (!rows.length) return;
    const header = cols.map(c => c.name).join(',');
    const body = rows.map(r => r.map(v => `"${(v ?? '').replace(/"/g, '""')}"`).join(',')).join('\n');
    const csv = header + '\n' + body;
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = 'query_result.csv'; a.click();
    URL.revokeObjectURL(url);
    toast({ title: 'CSV 已导出' });
  };

  const copyAsMarkdown = () => {
    if (!rows.length) return;
    const header = '| ' + cols.map(c => c.name).join(' | ') + ' |';
    const sep = '| ' + cols.map(() => '---').join(' | ') + ' |';
    const body = rows.map(r => '| ' + r.map(v => v ?? 'NULL').join(' | ') + ' |').join('\n');
    navigator.clipboard.writeText(header + '\n' + sep + '\n' + body);
    toast({ title: '已复制为 Markdown 表格' });
  };

  // ----- 右键菜单：复制操作 -----
  const copyText = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast({ title: '已复制' });
    } catch (e: any) {
      toast({ title: '复制失败', description: e?.message, variant: 'destructive' });
    }
  };

  const copyCell = (rowIdx: number, colIdx: number) => {
    const v = pagedRows[rowIdx]?.[colIdx];
    copyText(formatValue(v));
  };

  const copyRowTSV = (rowIdx: number) => {
    const row = pagedRows[rowIdx];
    if (!row) return;
    copyText(row.map(formatValue).join('\t'));
  };

  const copyRowInsert = (rowIdx: number) => {
    const row = pagedRows[rowIdx];
    if (!row) return;
    const colList = cols.map(c => `\`${c.name}\``).join(', ');
    const valList = row.map((v, i) => toSqlLiteral(v, cols[i].databaseType)).join(', ');
    const sql = `INSERT INTO \`${resolvedTableName}\` (${colList}) VALUES (${valList});`;
    copyText(sql);
  };

  const copyColName = (colIdx: number) => {
    const c = cols[colIdx];
    if (!c) return;
    copyText(c.name);
  };

  /**
   * 渲染单元格内容。
   */
  const renderCell = (value: string | null | undefined) => {
    if (value === null || value === undefined) {
      return <span className="text-muted-foreground italic">NULL</span>;
    }
    const str = value;

    // JSON 值仍以等宽字体显示，但同样不截断 textContent
    let isJson = false;
    if (str.startsWith('{') || str.startsWith('[')) {
      try {
        JSON.parse(str);
        isJson = true;
      } catch {}
    }

    return (
      <div
        className={`truncate ${isJson ? 'text-blue-400 font-mono' : ''}`}
        title={str}
      >
        {str}
      </div>
    );
  };

  // 生成分页页码按钮
  const renderPageButtons = () => {
    const buttons: (number | 'ellipsis')[] = [];
    const maxVisible = 5;

    if (totalPages <= maxVisible + 2) {
      // 全部显示
      for (let i = 1; i <= totalPages; i++) buttons.push(i);
    } else {
      buttons.push(1);
      const start = Math.max(2, currentPage - 1);
      const end = Math.min(totalPages - 1, currentPage + 1);
      if (start > 2) buttons.push('ellipsis');
      for (let i = start; i <= end; i++) buttons.push(i);
      if (end < totalPages - 1) buttons.push('ellipsis');
      buttons.push(totalPages);
    }

    return buttons.map((p, i) =>
      p === 'ellipsis' ? (
        <span key={`e${i}`} className="px-1 text-xs text-muted-foreground">...</span>
      ) : (
        <Button
          key={p}
          variant={p === currentPage ? 'default' : 'ghost'}
          size="sm"
          className="h-6 w-6 p-0 text-xs"
          onClick={() => setCurrentPage(p)}
        >
          {p}
        </Button>
      )
    );
  };

  return (
    <div className="flex flex-col h-full">
      {/* 工具栏 */}
      <div className="flex items-center gap-2 px-2 py-1 border-b border-border shrink-0 text-xs text-muted-foreground">
        <span>{rows.length} 行</span>
        <span>·</span>
        <span>{cols.length} 列</span>
        {(result.durationMs ?? 0) > 0 && <><span>·</span><span>{result.durationMs}ms</span></>}
        {result.truncated && <><span>·</span><span className="text-amber-500">结果已截断</span></>}
        <div className="flex-1" />
        <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={exportCSV}>
          <Download className="h-3 w-3 mr-1" />CSV
        </Button>
        <Button variant="ghost" size="sm" className="h-6 text-xs" onClick={copyAsMarkdown}>
          <Copy className="h-3 w-3 mr-1" />Markdown
        </Button>
      </div>

      {/* 表格 */}
      <div ref={parentRef} className="flex-1 overflow-auto">
        <table
          className="text-xs"
          style={{
            tableLayout: 'fixed',
            width: ROW_INDEX_WIDTH + cols.reduce((sum, c) => sum + getColWidth(c.name), 0),
            minWidth: '100%',
          }}
        >
          <colgroup>
            <col style={{ width: ROW_INDEX_WIDTH }} />
            {cols.map((col, i) => (
              <col key={i} style={{ width: getColWidth(col.name) }} />
            ))}
          </colgroup>
          <thead className="sticky top-0 bg-muted/80 backdrop-blur-sm z-10">
            <tr>
              <th className="px-2 py-1 text-left font-medium text-muted-foreground">#</th>
              {cols.map((col, i) => (
                <ContextMenu key={i}>
                  <ContextMenuTrigger asChild>
                    <th className="relative px-2 py-1 text-left font-medium text-muted-foreground whitespace-nowrap cursor-default overflow-hidden">
                      <div className="truncate pr-2" title={`${col.name} ${col.databaseType}`}>
                        {col.name}
                        <span className="ml-1 opacity-50">{col.databaseType}</span>
                      </div>
                      {/* 拖拽手柄：右边缘 */}
                      <div
                        role="separator"
                        aria-orientation="vertical"
                        onMouseDown={(e) => onResizeMouseDown(e, col.name)}
                        onDoubleClick={(e) => onResizeDoubleClick(e, col.name)}
                        title="拖拽调整列宽，双击重置"
                        className="absolute top-0 right-0 h-full w-1.5 cursor-col-resize select-none hover:bg-primary/40 active:bg-primary/60 z-20"
                      />
                    </th>
                  </ContextMenuTrigger>
                  <ContextMenuContent>
                    <ContextMenuItem onClick={() => copyColName(i)}>
                      <Hash className="h-3 w-3" />
                      复制列名
                    </ContextMenuItem>
                  </ContextMenuContent>
                </ContextMenu>
              ))}
            </tr>
          </thead>
          <tbody>
            {(() => {
              const virtualItems = virtualizer.getVirtualItems();
              const totalSize = virtualizer.getTotalSize();
              const paddingTop = virtualItems.length > 0 ? virtualItems[0].start : 0;
              const paddingBottom =
                virtualItems.length > 0
                  ? totalSize - virtualItems[virtualItems.length - 1].end
                  : 0;

              return (
                <>
                  {paddingTop > 0 && (
                    <tr aria-hidden="true">
                      <td colSpan={cols.length + 1} style={{ height: paddingTop, padding: 0, border: 0 }} />
                    </tr>
                  )}
                  {virtualItems.map(virtualRow => {
                    const rowIdx = virtualRow.index;
                    const row = pagedRows[rowIdx];
                    return (
                      <ContextMenu key={rowIdx}>
                        <ContextMenuTrigger asChild>
                          <tr
                            className="border-b border-border/50 hover:bg-accent/30"
                            style={{ height: virtualRow.size }}
                          >
                            <td className="px-2 py-0.5 text-muted-foreground overflow-hidden">{offset + rowIdx + 1}</td>
                            {row?.map((cell, ci) => (
                              <ContextMenu key={ci}>
                                <ContextMenuTrigger asChild>
                                  <td className="px-2 py-0.5 align-middle overflow-hidden">{renderCell(cell)}</td>
                                </ContextMenuTrigger>
                                <ContextMenuContent>
                                  <ContextMenuItem onClick={() => copyCell(rowIdx, ci)}>
                                    <Copy className="h-3 w-3" />
                                    复制单元格
                                  </ContextMenuItem>
                                  <ContextMenuSeparator />
                                  <ContextMenuItem onClick={() => copyRowTSV(rowIdx)}>
                                    <Clipboard className="h-3 w-3" />
                                    复制行 (TSV)
                                  </ContextMenuItem>
                                  <ContextMenuItem onClick={() => copyRowInsert(rowIdx)}>
                                    <FileText className="h-3 w-3" />
                                    复制行 (INSERT SQL)
                                  </ContextMenuItem>
                                </ContextMenuContent>
                              </ContextMenu>
                            ))}
                          </tr>
                        </ContextMenuTrigger>
                        <ContextMenuContent>
                          <ContextMenuItem onClick={() => copyRowTSV(rowIdx)}>
                            <Clipboard className="h-3 w-3" />
                            复制行 (TSV)
                          </ContextMenuItem>
                          <ContextMenuItem onClick={() => copyRowInsert(rowIdx)}>
                            <FileText className="h-3 w-3" />
                            复制行 (INSERT SQL)
                          </ContextMenuItem>
                        </ContextMenuContent>
                      </ContextMenu>
                    );
                  })}
                  {paddingBottom > 0 && (
                    <tr aria-hidden="true">
                      <td colSpan={cols.length + 1} style={{ height: paddingBottom, padding: 0, border: 0 }} />
                    </tr>
                  )}
                </>
              );
            })()}
          </tbody>
        </table>
      </div>

      {/* 分页栏 */}
      {showPagination && (
        <div className="flex items-center gap-2 px-2 py-1 border-t border-border shrink-0 text-xs text-muted-foreground">
          {/* 左侧：总行数 + 每页行数 */}
          <span>共 {rows.length} 行</span>
          <span>·</span>
          <Select value={String(pageSize)} onValueChange={v => { setPageSize(Number(v)); setCurrentPage(1); }}>
            <SelectTrigger className="h-6 w-[95px] text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {PAGE_SIZE_OPTIONS.map(s => (
                <SelectItem key={s} value={String(s)}>{s} 行/页</SelectItem>
              ))}
            </SelectContent>
          </Select>

          <div className="flex-1" />

          {/* 右侧：翻页按钮 */}
          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => setCurrentPage(1)} disabled={currentPage === 1}>
            <ChevronsLeft className="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => setCurrentPage(p => Math.max(1, p - 1))} disabled={currentPage === 1}>
            <ChevronLeft className="h-3 w-3" />
          </Button>

          {renderPageButtons()}

          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))} disabled={currentPage === totalPages}>
            <ChevronRight className="h-3 w-3" />
          </Button>
          <Button variant="ghost" size="sm" className="h-6 w-6 p-0" onClick={() => setCurrentPage(totalPages)} disabled={currentPage === totalPages}>
            <ChevronsRight className="h-3 w-3" />
          </Button>
        </div>
      )}
    </div>
  );
};

export default ResultPanel;
