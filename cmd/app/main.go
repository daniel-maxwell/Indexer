package main

import (
    "context"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "indexer/internal/config"
    "indexer/internal/logger"
    "indexer/internal/pkg/administrator"
    "go.uber.org/zap"
)

func main() {
    // 1. Load Configuration
    config, err := config.LoadConfig()
    if err != nil {
        fmt.Printf("Failed to load config: %v\n", err)
        os.Exit(1)
    }

    // 2. Initialize logger
    if err := logger.InitLogger(config.LogLevel); err != nil {
        fmt.Printf("Failed to initialize logger: %v\n", err)
        os.Exit(1)
    }
    defer logger.Log.Sync() // flush any buffered log entries

    logger.Log.Info("Starting indexer service", zap.String("version", "1.0.0"))

    // Construct the administrator with config
    admin := administrator.New(config)

    // Create a cancellable context for graceful shutdown
    context, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start processing in the background
    if err := admin.ProcessAndIndex(context); err != nil {
        logger.Log.Fatal("Failed to start indexer processing", zap.Error(err))
    }

    // Start ingest service (for receiving page data) in another goroutine
    go func() {
        admin.StartService(config.ServerPort)
    }()

    // Listen for OS signals to gracefully shut down
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    select {
    case s := <-sigChan:
        logger.Log.Info("Received shutdown signal", zap.String("signal", s.String()))
        cancel()
    }

    // Give some time for final cleanup, if needed
    time.Sleep(2 * time.Second)
    logger.Log.Info("Indexer shutdown complete")
}



