package http

import (
	"context"
	"sync"

	"cyancat/internal/infra/api"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// FileDialogAPI 文件选择 API（暴露给前端）
type FileDialogAPI struct {
	mu  sync.RWMutex
	ctx context.Context
}

func NewFileDialogAPI() *FileDialogAPI {
	return &FileDialogAPI{}
}

func SetFileDialogAPIContext(api *FileDialogAPI, ctx context.Context) {
	if api == nil {
		return
	}
	api.mu.Lock()
	defer api.mu.Unlock()
	api.ctx = ctx
}

func (a *FileDialogAPI) SelectSQLiteDatabaseFile() *api.Response[string] {
	a.mu.RLock()
	ctx := a.ctx
	a.mu.RUnlock()
	if ctx == nil {
		return api.Fail[string](api.ErrorCode, "", "wails context is not ready")
	}

	path, err := runtime.OpenFileDialog(ctx, runtime.OpenDialogOptions{
		Title: "选择 SQLite 数据库文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "SQLite Database (*.db;*.sqlite;*.sqlite3)", Pattern: "*.db;*.sqlite;*.sqlite3"},
			{DisplayName: "All Files (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return api.Fail[string](api.ErrorCode, "", err.Error())
	}
	return api.Success(path)
}
