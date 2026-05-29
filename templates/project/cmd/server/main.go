//go:build ignore

package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"{{.ModuleName}}/internal/platform/app"
	"{{.ModuleName}}/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 先装配平台能力，再注册业务路由，把初始化失败前置到启动期。
	serverApp, err := app.NewServer(ctx, app.ServerOptions{
		RegisterRoutes: server.RegisterRoutes,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = serverApp.Close(context.Background())
	}()

	if err := serverApp.Run(ctx); err != nil {
		log.Fatal(err)
	}
}
