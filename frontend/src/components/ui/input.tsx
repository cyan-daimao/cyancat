import * as React from "react"

import { cn } from "@/lib/utils"

/** Input 输入框属性接口 */
export interface InputProps
  extends React.InputHTMLAttributes<HTMLInputElement> {}

/** Input 输入框组件 */
const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        // macOS WKWebView 默认启用系统级自动大写/自动纠正，
        // 会破坏连接名、主机、表名等技术标识符的输入，统一关闭（调用方可显式覆盖）
        autoCapitalize="off"
        autoCorrect="off"
        spellCheck={false}
        className={cn(
          "flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
          className
        )}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = "Input"

/** Input 输入框组件 */
export { Input }
