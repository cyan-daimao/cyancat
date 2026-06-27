package http

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cyancat/internal/adapter/dto"
	"cyancat/internal/infra/api"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExportAPI 文件导出 API（暴露给前端）
type ExportAPI struct {
	mu  sync.RWMutex
	ctx context.Context
}

// NewExportAPI 创建 ExportAPI
func NewExportAPI() *ExportAPI {
	return &ExportAPI{}
}

// SetExportAPIContext 注入 Wails runtime context
func SetExportAPIContext(api *ExportAPI, ctx context.Context) {
	if api == nil {
		return
	}
	api.setContext(ctx)
}

func (a *ExportAPI) setContext(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.ctx = ctx
}

// ExportCSV 选择目录并保存 CSV 文件
func (a *ExportAPI) ExportCSV(req *dto.ExportCSVRequest) *api.Response[*dto.ExportCSVResult] {
	if req == nil {
		return api.Fail[*dto.ExportCSVResult](api.BadRequestCode, nil, "request cannot be nil")
	}

	ctx := a.context()
	if ctx == nil {
		return api.Fail[*dto.ExportCSVResult](api.ErrorCode, nil, "wails context is not ready")
	}

	dir, err := runtime.OpenDirectoryDialog(ctx, runtime.OpenDialogOptions{
		Title:                "选择 CSV 保存目录",
		CanCreateDirectories: true,
	})
	if err != nil {
		return api.Fail[*dto.ExportCSVResult](api.ErrorCode, nil, err.Error())
	}
	if dir == "" {
		return api.Success(&dto.ExportCSVResult{Saved: false})
	}

	filename := normalizeCSVFilename(req.Filename)
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(req.Content), 0644); err != nil {
		return api.Fail[*dto.ExportCSVResult](api.ErrorCode, nil, err.Error())
	}

	return api.Success(&dto.ExportCSVResult{
		Saved: true,
		Path:  path,
	})
}

func (a *ExportAPI) context() context.Context {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.ctx
}

func normalizeCSVFilename(filename string) string {
	name := strings.TrimSpace(filename)
	if name == "" {
		name = "query_result.csv"
	}
	name = filepath.Base(name)
	if name == "." || name == ".." || name == string(filepath.Separator) {
		name = "query_result.csv"
	}
	if strings.ToLower(filepath.Ext(name)) != ".csv" {
		name += ".csv"
	}
	return name
}
