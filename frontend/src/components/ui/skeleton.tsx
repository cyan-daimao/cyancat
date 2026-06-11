import * as React from "react"

import { cn } from "@/lib/utils"

/** Skeleton 骨架屏组件 - 加载状态的占位容器 */
function Skeleton({
  className,
  ...props
}: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div
      className={cn("animate-pulse rounded-md bg-muted", className)}
      {...props}
    />
  )
}

export { Skeleton }
