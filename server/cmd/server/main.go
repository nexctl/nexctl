package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/nexctl/nexctl/server/internal/app"
)

// loadDotEnv 加载当前目录与上一级的 .env（便于在 server/ 下 go run 时使用仓库根目录配置）。
// 不覆盖已在环境中显式设置的变量（godotenv 默认行为）。
func loadDotEnv() {
	_ = godotenv.Load(".env")
	_ = godotenv.Load("../.env")
}

// main starts the NexCtl control plane server.
func main() {
	loadDotEnv()
	configPath := flag.String("config", "configs/config.example.yaml", "server config path")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	application, err := app.New(ctx, *configPath)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}

	if err := application.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
