//go:build ignore

package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"{{.ModuleName}}/internal/platform/config"
	"{{.ModuleName}}/internal/platform/llm"
	"{{.ModuleName}}/internal/platform/logging"
	"{{.ModuleName}}/internal/platform/storage"
	"{{.ModuleName}}/internal/platform/task"
	redisstream "{{.ModuleName}}/internal/platform/task/backend/queue/redisstream"
	taskmongo "{{.ModuleName}}/internal/platform/task/backend/storage/mongo"
)

type WorkerOptions struct {
	RegisterTasks func(a *WorkerApp) error
}

type WorkerApp struct {
	Config   config.Config
	Logger   *slog.Logger
	Redis    redis.UniversalClient
	Mongo    *mongo.Client
	Registry *task.Registry
	Worker   *task.Worker
	LLM      *llm.Client
	logBoot  *logging.Bootstrap
}

func NewWorker(ctx context.Context, opts WorkerOptions) (*WorkerApp, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logBoot, err := logging.New(cfg.Log, "worker")
	if err != nil {
		return nil, err
	}

	app := &WorkerApp{
		Config:   cfg,
		Logger:   logBoot.Logger,
		logBoot:  logBoot,
		Registry: task.NewRegistry(),
	}

	// Worker 角色必须消费任务，因此要求任务能力显式启用。
	if !cfg.Task.Enabled {
		_ = app.Close(context.Background())
		return nil, fmt.Errorf("worker requires TASK_ENABLED=true")
	}

	if cfg.Redis.Enabled {
		app.Redis, err = storage.NewRedis(cfg.Redis)
		if err != nil {
			_ = app.Close(context.Background())
			return nil, err
		}
	}
	if cfg.Mongo.Enabled {
		app.Mongo, err = storage.NewMongo(cfg.Mongo)
		if err != nil {
			_ = app.Close(context.Background())
			return nil, err
		}
	}

	if app.Mongo == nil || app.Redis == nil {
		_ = app.Close(context.Background())
		return nil, fmt.Errorf("worker requires redis and mongo when task is enabled")
	}

	collection := storage.TaskCollection(app.Mongo, cfg.Mongo.Database)
	repo, err := taskmongo.NewRepository(ctx, collection)
	if err != nil {
		_ = app.Close(context.Background())
		return nil, err
	}
	queue := redisstream.NewQueue(app.Redis, cfg.Task, app.Logger)
	app.Worker = task.NewWorker(repo, queue, app.Registry, app.Logger, cfg.Task)

	if cfg.LLM.Enabled {
		app.LLM, err = llm.New(cfg.LLM)
		if err != nil {
			_ = app.Close(context.Background())
			return nil, err
		}
	}

	if opts.RegisterTasks == nil {
		_ = app.Close(context.Background())
		return nil, fmt.Errorf("RegisterTasks is required")
	}
	if err := opts.RegisterTasks(app); err != nil {
		_ = app.Close(context.Background())
		return nil, err
	}

	return app, nil
}

func (a *WorkerApp) Run(ctx context.Context) error {
	a.Logger.Info("worker starting",
		"consumer_group", a.Config.Task.ConsumerGroup,
		"consumer_name", a.Config.Task.ConsumerName,
		"worker_concurrency", a.Config.Task.WorkerConcurrency,
	)
	err := a.Worker.Run(ctx)
	a.Logger.Info("worker stopped")
	return err
}

func (a *WorkerApp) Close(ctx context.Context) error {
	var firstErr error
	if err := storage.CloseMongo(a.Mongo); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := storage.CloseRedis(a.Redis); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := a.logBoot.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	_ = ctx
	return firstErr
}
