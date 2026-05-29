package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/zyq/anvil/internal/agent"
	"github.com/zyq/anvil/internal/platform/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	workerApp, err := app.NewWorker(ctx, app.WorkerOptions{
		RegisterTasks: agent.RegisterTaskHandlers,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = workerApp.Close(context.Background())
	}()

	if err := workerApp.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
