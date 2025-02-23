package main

import (
	"context"
	"fmt"
	"indexer/internal/pkg/administrator"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

    // Construct our administrator.
    admin := administrator.New()

    // Create a cancellable context so we can gracefully shut down.
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start the administrator in a background goroutine.
    if err := admin.ProcessAndIndex(ctx); err != nil {
        log.Fatalf("failed to start indexer processing: %v", err)
    }

    // Listen for OS signals to gracefully shut down.
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    select {
    case s := <-sigChan:
        fmt.Printf("received signal %s, shutting down...\n", s)
        cancel()
    }

    // Give some time for cleanup if needed
    time.Sleep(2 * time.Second)
    log.Println("indexer shutdown complete")
}


