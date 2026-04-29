package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/yourorg/cronwatch/internal/config"
)

const defaultConfigPath = "/etc/cronwatch/cronwatch.yaml"

func main() {
	configPath := flag.String("config", defaultConfigPath, "path to cronwatch YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cronwatch: failed to load config: %v\n", err)
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	slog.Info("cronwatch starting",
		"config", *configPath,
		"jobs", len(cfg.Jobs),
		"state_file", cfg.StateFile,
	)

	for _, job := range cfg.Jobs {
		slog.Debug("registered job",
			"name", job.Name,
			"schedule", job.Schedule,
			"max_duration", job.MaxDuration,
			"drift_threshold", job.DriftThreshold,
		)
	}

	// TODO: start scheduler and monitoring loop
	slog.Info("cronwatch initialised — scheduler not yet implemented")
}
