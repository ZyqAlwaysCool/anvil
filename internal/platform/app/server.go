package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"

	"github.com/ZyqAlwaysCool/anvil/internal/platform/config"
	httpmiddleware "github.com/ZyqAlwaysCool/anvil/internal/platform/http/middleware"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/llm"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/logging"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/storage"
	"github.com/ZyqAlwaysCool/anvil/internal/platform/task"
	redisstream "github.com/ZyqAlwaysCool/anvil/internal/platform/task/backend/queue/redisstream"
	taskmongo "github.com/ZyqAlwaysCool/anvil/internal/platform/task/backend/storage/mongo"
)

type ServerOptions struct {
	RegisterRoutes func(a *ServerApp) error
}

type ServerApp struct {
	Config  config.Config
	Logger  *slog.Logger
	Engine  *gin.Engine
	Redis   redis.UniversalClient
	MySQL   *gorm.DB
	SQLite  *gorm.DB
	Mongo   *mongo.Client
	Task    *task.Manager
	LLM     *llm.Client
	logBoot *logging.Bootstrap
	httpSrv *http.Server
}

// NewServer 固定初始化顺序：config -> logging -> storage -> task(可选) -> llm(可选) -> gin -> routes。
func NewServer(ctx context.Context, opts ServerOptions) (*ServerApp, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logBoot, err := logging.New(cfg.Log, "server")
	if err != nil {
		return nil, err
	}

	app := &ServerApp{
		Config:  cfg,
		Logger:  logBoot.Logger,
		logBoot: logBoot,
	}

	if err := app.initStorage(ctx); err != nil {
		_ = app.Close(context.Background())
		return nil, err
	}
	if err := app.initTask(ctx); err != nil {
		_ = app.Close(context.Background())
		return nil, err
	}
	if err := app.initLLM(); err != nil {
		_ = app.Close(context.Background())
		return nil, err
	}
	if err := app.initHTTP(opts); err != nil {
		_ = app.Close(context.Background())
		return nil, err
	}

	return app, nil
}

func (a *ServerApp) initStorage(ctx context.Context) error {
	var err error
	if a.Config.Redis.Enabled {
		a.Redis, err = storage.NewRedis(a.Config.Redis)
		if err != nil {
			return err
		}
	}
	if a.Config.MySQL.Enabled {
		a.MySQL, err = storage.NewMySQL(a.Config.MySQL)
		if err != nil {
			return err
		}
	}
	if a.Config.SQLite.Enabled {
		a.SQLite, err = storage.NewSQLite(a.Config.SQLite)
		if err != nil {
			return err
		}
	}
	if a.Config.Mongo.Enabled {
		a.Mongo, err = storage.NewMongo(a.Config.Mongo)
		if err != nil {
			return err
		}
	}
	_ = ctx
	return nil
}

func (a *ServerApp) initTask(ctx context.Context) error {
	// 任务系统是可选能力：只有显式启用且 Redis/Mongo 均已装配时才初始化 Manager。
	if !a.Config.Task.Enabled {
		a.Logger.Info("task system disabled, skip task manager init")
		return nil
	}
	if a.Redis == nil || a.Mongo == nil {
		return fmt.Errorf("task enabled but redis or mongo is not available")
	}

	collection := storage.TaskCollection(a.Mongo, a.Config.Mongo.Database)
	repo, err := taskmongo.NewRepository(ctx, collection)
	if err != nil {
		return err
	}
	queue := redisstream.NewQueue(a.Redis, a.Config.Task, a.Logger)
	a.Task = task.NewManager(repo, queue, a.Logger, a.Config.Task, "mongodb", "redis_stream")
	a.Logger.Info("task system initialized", "stream_key", a.Config.Task.StreamKey)
	return nil
}

func (a *ServerApp) initLLM() error {
	// LLM 关闭时返回 nil 客户端，不阻塞 HTTP 服务启动。
	client, err := llm.New(a.Config.LLM)
	if err != nil {
		return err
	}
	a.LLM = client
	if client == nil {
		a.Logger.Info("llm disabled, skip llm client init")
	}
	return nil
}

func (a *ServerApp) initHTTP(opts ServerOptions) error {
	gin.SetMode(a.Config.HTTP.GinMode)
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(httpmiddleware.TraceID())
	engine.Use(httpmiddleware.AccessLog(a.Logger))
	engine.Use(httpmiddleware.CORS(httpmiddleware.CORSOptions{
		AllowOrigins: a.Config.CORS.AllowOrigins,
		AllowMethods: a.Config.CORS.AllowMethods,
		AllowHeaders: a.Config.CORS.AllowHeaders,
		MaxAge:       a.Config.CORS.MaxAge,
	}))

	a.Engine = engine
	if opts.RegisterRoutes == nil {
		return fmt.Errorf("RegisterRoutes is required")
	}
	return opts.RegisterRoutes(a)
}

func (a *ServerApp) Run(ctx context.Context) error {
	addr := fmt.Sprintf(":%s", a.Config.HTTP.Port)
	fields := []any{"addr", addr, "gin_mode", a.Config.HTTP.GinMode}
	if a.Task != nil {
		fields = append(fields, "task_stream_key", a.Config.Task.StreamKey)
	}
	a.Logger.Info("server starting", fields...)

	a.httpSrv = &http.Server{Addr: addr, Handler: a.Engine}
	errCh := make(chan error, 1)
	go func() {
		err := a.httpSrv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
			return
		}
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.httpSrv.Shutdown(shutdownCtx); err != nil {
			return err
		}
		a.Logger.Info("server stopped")
		return <-errCh
	}
}

func (a *ServerApp) Close(ctx context.Context) error {
	var firstErr error
	if a.httpSrv != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		if err := a.httpSrv.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			firstErr = err
		}
	}
	if err := storage.CloseMongo(a.Mongo); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := storage.CloseSQLite(a.SQLite); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := storage.CloseMySQL(a.MySQL); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := storage.CloseRedis(a.Redis); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := a.logBoot.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}
