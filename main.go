package main

import (
	"context"
	"embed"
	"os"
	"path/filepath"

	"cyancat/internal/adapter"
	adapterhttp "cyancat/internal/adapter/http"
	"cyancat/internal/application/connectionservice"
	"cyancat/internal/application/queryservice"
	"cyancat/internal/application/schemaservice"
	"cyancat/internal/infra/db"
	"cyancat/internal/infra/db/connectionrepo"
	"cyancat/internal/infra/db/historyrepo"
	"cyancat/internal/infra/driver"
	redisdriver "cyancat/internal/infra/driver/redis"
	kafkadriver "cyancat/internal/infra/driver/kafka"
	mysqldriver "cyancat/internal/infra/driver/mysql"
	postgresdriver "cyancat/internal/infra/driver/postgres"
	starrocksdriver "cyancat/internal/infra/driver/starrocks"
	sqlitedriver "cyancat/internal/infra/driver/sqlite"
	"cyancat/internal/infra/eventbus"
	"cyancat/internal/infra/keychain"
	"cyancat/internal/infra/logger"
	"cyancat/internal/infra/session"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// 1. 初始化日志
	logger.Init("info", true)

	// 2. 注册数据库驱动
	driver.Register(mysqldriver.New())
	driver.Register(postgresdriver.New())
	driver.Register(sqlitedriver.New())
	driver.Register(starrocksdriver.New())
	driver.Register(kafkadriver.New())
	driver.Register(redisdriver.New())

	// 3. 初始化本地 SQLite
	dbPath := dbPath()
	if err := db.Init(dbPath); err != nil {
		logger.L().Fatal().Err(err).Msg("failed to init sqlite")
	}

	// 4. 自动迁移表结构
	if err := connectionrepo.AutoMigrate(); err != nil {
		logger.L().Fatal().Err(err).Msg("connection migration failed")
	}
	if err := historyrepo.AutoMigrate(); err != nil {
		logger.L().Fatal().Err(err).Msg("history migration failed")
	}

	// 5. 初始化 Keychain（V1.0 使用 AES 兜底）
	if err := keychain.Init(); err != nil {
		logger.L().Warn().Err(err).Msg("keychain init failed, using fallback mode")
	}

	// 6. 装配依赖：infra -> domain -> application -> adapter
	masterKey := getMasterKey()
	connectionRepository := connectionrepo.NewConnectionRepository(masterKey)
	sessionManager := session.NewManager()
	bus := eventbus.Default()

	connectionService := connectionservice.NewConnectionServiceImpl(connectionRepository, sessionManager)
	historyRepo := historyrepo.NewQueryHistoryRepository()
	queryService := queryservice.NewQueryServiceImpl(sessionManager, bus, historyRepo)
	schemaService := schemaservice.NewSchemaServiceImpl(sessionManager)

	app := adapter.NewApp(connectionService, queryService, schemaService)

	// 7. 启动 Wails
	if err := wails.Run(&options.App{
		Title:  "DBStudio (cyancat)",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			// 把 Wails ctx 注入 EventBus，让后端能向前端推流式数据
			eventbus.Init(ctx)
			adapterhttp.SetExportAPIContext(app.ExportAPI, ctx)
			adapterhttp.SetFileDialogAPIContext(app.FileDialogAPI, ctx)
			logger.L().Info().Msg("cyancat started")
		},
		OnShutdown: func(ctx context.Context) {
			// 应用退出时清理所有活跃数据库连接
			if err := sessionManager.CloseAll(); err != nil {
				logger.L().Warn().Err(err).Msg("close sessions on shutdown")
			}
			logger.L().Info().Msg("cyancat shutdown")
			// 刷新并关闭日志文件，保证 Windows GUI 子系统下日志完整落盘
			logger.Close()
		},
		Bind: []interface{}{
			app,
			app.ConnectionAPI,
			app.QueryAPI,
			app.SchemaAPI,
			app.ExportAPI,
			app.FileDialogAPI,
		},
	}); err != nil {
		logger.L().Fatal().Err(err).Msg("wails run failed")
	}
}

// dbPath 返回本地 SQLite 数据库路径
func dbPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "cyancat.db"
	}
	dir := filepath.Join(home, ".cyancat")
	_ = os.MkdirAll(dir, 0700)
	return filepath.Join(dir, "cyancat.db")
}

// getMasterKey 从 keychain 获取 AES 主密钥
// V1.0 使用固定 32 字节 key（从环境变量或文件读取），V2.0 改为 PBKDF2 派生
func getMasterKey() []byte {
	// 优先从环境变量获取
	keyHex := os.Getenv("CYANCAT_MASTER_KEY")
	if keyHex != "" && len(keyHex) == 64 {
		key := make([]byte, 32)
		for i := 0; i < 32; i++ {
			high := hexVal(keyHex[i*2])
			low := hexVal(keyHex[i*2+1])
			if high < 0 || low < 0 {
				break
			}
			key[i] = byte(high<<4 | low)
		}
		return key
	}

	// 尝试从文件读取
	home, err := os.UserHomeDir()
	if err == nil {
		keyData, err := os.ReadFile(filepath.Join(home, ".cyancat", "master.key"))
		if err == nil && len(keyData) == 32 {
			return keyData
		}
	}

	// 兜底：生成默认 key（仅开发环境，生产环境必须配置）
	logger.L().Warn().Msg("using default master key - configure CYANCAT_MASTER_KEY or ~/.cyancat/master.key for production")
	return []byte("cyancat-default-key-change-me-32!")[:32]
}

func hexVal(c byte) int {
	switch {
	case '0' <= c && c <= '9':
		return int(c - '0')
	case 'a' <= c && c <= 'f':
		return int(c - 'a' + 10)
	case 'A' <= c && c <= 'F':
		return int(c - 'A' + 10)
	default:
		return -1
	}
}
