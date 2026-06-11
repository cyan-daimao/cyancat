// Package logger 提供结构化日志功能
package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var log zerolog.Logger

// Init 初始化全局日志
func Init(level string, pretty bool) {
	var output io.Writer = os.Stderr

	if pretty {
		output = zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.DateTime,
		}
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	log = zerolog.New(output).
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