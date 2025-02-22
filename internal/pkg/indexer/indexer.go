package indexer

import (
    "context"
    "fmt"
	"log"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/queue"
)

// Indexer interface defines the methods that an indexer should implement.
type Indexer interface {
    EnqueuePageData(ctx context.Context, data models.PageData) error
    StartProcessing(ctx context.Context) error
}

// implementation of the Indexer interface.
type indexer struct {
    // queue is the in-memory slice-based queue (already implemented).
    queue *queue.Queue

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
    }
}

func (i *indexer) EnqueuePageData(ctx context.Context, data models.PageData) error {
    // This quickly returns so the crawler can move on.
    return i.queue.Insert(data)
}

// Will eventually handle reading items from the queue and passing them to the processing pipeline.
func (i *indexer) StartProcessing(ctx context.Context) error {
    go func() {
        for {
            select {
            case <-ctx.Done():
                fmt.Println("context canceled, stopping indexer processing")
                return
            default:
                // In a real implementation, we'd Dequeue and process each item
                // (do cleanup, NLP, dedupe, spam check, etc.)
                // For Step 1, we leave this as a placeholder.
            }
        }
    }()
    return nil
}
