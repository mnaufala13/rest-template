package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mnaufala13/rest-template/config"
	"github.com/mnaufala13/rest-template/http"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type sConfig struct {
	Log  config.Log          `json:"log"`
	Http config.ServerConfig `json:"http"`
}

func loadConfig(path string) sConfig {
	content, err := os.ReadFile(path)
	if err != nil {
		slog.Error(fmt.Sprintf("read config: %v", err))
		os.Exit(1)
	}
	cfg := sConfig{}
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		slog.Error(fmt.Sprintf("json unmarshal config: %v", err))
		os.Exit(1)
	}

	return cfg
}

func main() {
	// create context & handle termination signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer stop()

	args := os.Args
	if len(args) < 2 {
		slog.Error("arg config path can't empty")
		os.Exit(1)
	}
	path := args[1]
	cfg := loadConfig(path)
	setupLogger(cfg.Log.Severity, cfg.Log.Handler)

	dep := http.ServerDependency{
		CreateUser: func(ctx context.Context, username string) error {
			slog.Info(fmt.Sprintf("user %s created", username))
			return nil
		},
	}
	go http.Start(cfg.Http, dep)
	defer http.Stop(time.Duration(cfg.Http.ShutdownDelay), time.Duration(cfg.Http.ShutdownTimeout))

	<-ctx.Done()
}
