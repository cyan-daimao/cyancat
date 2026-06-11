// Package eventbus 封装 Wails 事件推送，用于后端向前端推送流式数据（如查询结果分批、执行进度）
package eventbus

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 事件名常量
const (
	// EventQueryRows 查询结果分批推送：data = {connID, sql, rows, columns, done}
	EventQueryRows = "query:rows"
	// EventQueryDone 查询完成（含统计信息）：data = {connID, sql, rowCount, duration}
	EventQueryDone = "query:done"
	// EventQueryError 查询错误：data = {connID, sql, error}
	EventQueryError = "query:error"
	// EventConnectionStateChanged 连接状态变化：data = {connID, state}
	EventConnectionStateChanged = "connection:state"
)

// Bus 事件总线接口
type Bus interface {
	// Emit 向前端推送事件
	Emit(eventName string, data ...any)
	// On 监听前端事件（V1.5 用）
	On(eventName string, handler func(args ...any))
}

// wailsBus 基于 Wails runtime 的事件总线实现
type wailsBus struct {
	mu  sync.RWMutex
	ctx context.Context
}

// defaultBus 全局默认事件总线
var defaultBus = &wailsBus{}

// Init 由 main.go 在 Wails OnStartup 中调用，注入 ctx
func Init(ctx context.Context) {
	defaultBus.mu.Lock()
	defaultBus.ctx = ctx
	defaultBus.mu.Unlock()
}

// Default 返回全局事件总线
func Default() Bus {
	return defaultBus
}

// Emit 推送事件给前端
func (b *wailsBus) Emit(eventName string, data ...any) {
	b.mu.RLock()
	ctx := b.ctx
	b.mu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.EventsEmit(ctx, eventName, data...)
}

// On 监听前端事件
func (b *wailsBus) On(eventName string, handler func(args ...any)) {
	b.mu.RLock()
	ctx := b.ctx
	b.mu.RUnlock()
	if ctx == nil {
		return
	}
	runtime.EventsOn(ctx, eventName, handler)
}
