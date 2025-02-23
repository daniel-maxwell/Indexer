package indexer

import (
    "context"
    "fmt"
	"log"
	"time"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/queue"
)

// Indexer interface defines the methods that an indexer should implement.
type Indexer interface {
    EnqueuePageData(ctx context.Context, data models.PageData) error
    StartProcessing(ctx context.Context) error
}

// indexer is an implementation of the Indexer interface.
type indexer struct {
    // queue is the in-memory slice-based queue (already implemented).
    queue *queue.Queue
	processor Processor

    // other configs, e.g. concurrency or ES connection strings, might go here.
}

// New creates a new instance of an indexer.
func New() Indexer {
	// Create/initialize the queue.
	pageQueue, err := queue.CreateQueue(1000)
	if err != nil {
		log.Fatalf("failed to create queue: %v", err)
	}
    return &indexer{
        queue: pageQueue,
		processor: NewProcessor(),
    }
}

func (i *indexer) EnqueuePageData(ctx context.Context, data models.PageData) error {
    // This quickly returns so the crawler can move on.
    return i.queue.Insert(data)
}

// Will eventually handle reading items from the queue and passing them to the processing pipeline.
func (i *indexer) StartProcessing(ctx context.Context) error {
    // For now, run in a single goroutine. 
    // In future this will spawn multiple worker goroutines for concurrency.
    go func() {
        for {
            select {
            case <-ctx.Done():
                fmt.Println("context canceled, stopping indexer processing")
                return
            default:
                // Attempt to grab an item from the queue
                item, err := i.queue.Remove()
                if err != nil {
                    // If the queue is empty, we can sleep briefly to avoid busy-looping
                    time.Sleep(200 * time.Millisecond)
                    continue
                }

                // Next, run our HTML cleanup & URL normalization
                cleaned, err := i.processor.CleanAndNormalize(item)
                if err != nil {
                    log.Printf("error cleaning/normalizing: %v", err)
                    continue
                }

                // For now, just log the results. 
                log.Printf("Cleaned & normalized item: %+v\n", cleaned)

                // Additional pipeline steps (next increments).
            }
        }
    }()
    return nil
}
