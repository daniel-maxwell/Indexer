package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    "time"
    "indexer/internal/pkg/config"
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/administrator"
    "go.uber.org/zap"
)

/**
This application requires a Redis instance to be running.
To start a Redis instance with Docker, run: docker run -p 6379:6379 --name redis -d redis:6.2
*/

func main() {
    config, err := config.LoadConfig()
    if err != nil {
        logger.Log.Error("Failed to load config", zap.Error(err))
        os.Exit(1)
    }

    if err := logger.InitLogger(config.LogLevel); err != nil {
        logger.Log.Error("Failed to initialize logger", zap.Error(err))
        os.Exit(1)
    }
    defer logger.Log.Sync()

    logger.Log.Info("Starting indexer service", zap.String("version", "1.0.0"))

    // Construct the administrator with config
    admin := administrator.New(config)

    // Create a cancellable context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start background processing
    if err := admin.ProcessAndIndex(ctx); err != nil {
        logger.Log.Fatal("Failed to start indexer processing", zap.Error(err))
    }

    // Start ingestion service in separate goroutine
    go func() {
        admin.StartService(config.ServerPort)
    }()

    // Listen for OS signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    s := <-sigChan
    logger.Log.Info("Received shutdown signal", zap.String("signal", s.String()))
    cancel() // stop reading from queue

    // Create a context with a timeout for shutdown
    _, shutdownCancel := context.WithTimeout(context.Background(), 30 * time.Second)
    defer shutdownCancel()

    // Gracefully stop the workers and BulkIndexer
    admin.Stop()

    logger.Log.Info("Shutdown complete")
}