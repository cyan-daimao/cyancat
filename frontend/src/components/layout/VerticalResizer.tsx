import React from 'react';
import { cn } from '@/lib/utils';

interface VerticalResizerProps {
  onResize: (deltaX: number) => void;
  onResizeStart?: () => void;
  onResizeEnd?: () => void;
  onDoubleClick?: () => void;
  disabled?: boolean;
}

const VerticalResizer: React.FC<VerticalResizerProps> = ({
  onResize,
  onResizeStart,
  onResizeEnd,
  onDoubleClick,
  disabled,
}) => {
  const handleMouseDown = (e: React.MouseEvent) => {
    if (disabled) return;
    e.preventDefault();
    const startX = e.clientX;
    onResizeStart?.();

    const prevCursor = document.body.style.cursor;
    const prevSelect = document.body.style.userSelect;
    document.body.style.cursor = 'col-resize';
    document.body.style.userSelect = 'none';

    const onMove = (ev: MouseEvent) => {
      onResize(ev.clientX - startX);
    };
    const onUp = () => {
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
      document.body.style.cursor = prevCursor;
      document.body.style.userSelect = prevSelect;
      onResizeEnd?.();
    };

    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  };

  return (
    <div
      role="separator"
      aria-orientation="vertical"
      onMouseDown={handleMouseDown}
      onDoubleClick={disabled ? undefined : onDoubleClick}
      className={cn(
        'group relative w-px shrink-0 bg-border transition-colors',
        disabled ? 'cursor-default opacity-50' : 'cursor-col-resize hover:bg-primary/40 active:bg-primary/60',
      )}
    >
      {/* 加宽不可见命中区域，便于拖拽 */}
      <div className="absolute inset-y-0 -left-1 -right-1" />
    </div>
  );
};

export default VerticalResizer;
