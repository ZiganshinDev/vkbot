package main

import (
	"log"
	"os"

	"github.com/ZiganshinDev/scheduleVKBot/internal/config"
	"github.com/ZiganshinDev/scheduleVKBot/internal/lib/logger/sl"
	"github.com/ZiganshinDev/scheduleVKBot/internal/service/vkbot"
	"github.com/ZiganshinDev/scheduleVKBot/internal/storage/sqlite"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("error loading .env file")
	}

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"Starting vkbot",
		slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	bot, err := vkbot.New()
	if err != nil {
		log.Error("failed to init vkbot", sl.Err(err))
		os.Exit(1)
	}

	if err := bot.Start(storage); err != nil {
		log.Error("failed to start vkbot", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
