package main

import (
	"context"
	"fmt"
	"indexer/internal/pkg/indexer"
	"indexer/internal/pkg/models"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {

    // Construct our indexer
    idx := indexer.New()

    // Create a cancellable context so we can gracefully shut down.
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start the indexer in a background goroutine.
    if err := idx.StartProcessing(ctx); err != nil {
        log.Fatalf("failed to start indexer processing: %v", err)
    }

    // In future, the indexer will accept incoming HTTP requests here that call idx.EnqueuePageData.
    // For now, let's just demonstrate how we might do one Enqueue:
    go func() {
        testData := fakePageData()
        if err := idx.EnqueuePageData(ctx, testData); err != nil {
            log.Printf("failed to enqueue test data: %v", err)
        } else {
            log.Println("successfully enqueued test data")
        }
    }()

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

func fakePageData() models.PageData{
    return models.PageData{
		URL:             "https://example.com",
		CanonicalURL:    "https://example.com",
		Title:           "Example Domain",
		Charset:         "UTF-8",
		MetaDescription: "This is an example domain used for illustrative examples in documents.",
		MetaKeywords:    "example, domain",
    }
}
