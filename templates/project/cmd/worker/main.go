//go:build ignore

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"anvil-scaffold-template/internal/agent"
	"anvil-scaffold-template/internal/platform/app"
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
