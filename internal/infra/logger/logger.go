// Package logger 提供结构化日志功能
package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var (
	log     zerolog.Logger
	logFile *os.File
)

// Init 初始化全局日志
//
// 日志同时输出到 stderr 和文件（~/.cyancat/cyancat.log）。
// Windows 下 Wails 编译为 GUI 子系统（链接器 -H windowsgui），
// stderr 无可见控制台，任何 Fatal 都会静默退出；写一份到文件
// 可保证启动期错误有迹可循。文件以 JSON 行格式追加写入。
func Init(level string, pretty bool) {
	var writers []io.Writer

	// stderr：开发期 pretty 控制台，否则纯 JSON
	if pretty {
		writers = append(writers, zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.DateTime,
		})
	} else {
		writers = append(writers, os.Stderr)
	}

	// 文件：始终追加 JSON 行格式，便于机器解析与事后排查
	if f := openLogFile(); f != nil {
		logFile = f
		writers = append(writers, f)
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	log = zerolog.New(io.MultiWriter(writers...)).
		Level(lvl).
		With().
		Timestamp().
		Caller().
		Logger()
}

// L 返回全局日志实例
func L() *zerolog.Logger {
	return &log
}

// Close 关闭日志文件，应用退出时调用
func Close() {
	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
}

// openLogFile 打开 ~/.cyancat/cyancat.log 用于追加写入。
// 失败时返回 nil，调用方应继续仅用 stderr，不阻断启动。
func openLogFile() *os.File {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	dir := filepath.Join(home, ".cyancat")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil
	}
	f, err := os.OpenFile(filepath.Join(dir, "cyancat.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil
	}
	return f
}
