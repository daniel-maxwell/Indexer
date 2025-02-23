package indexer

import (
    "context"
	"log"
	"time"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/queue"
	"indexer/internal/pkg/processor"
)

// Indexer interface defines the methods that an indexer should implement.
type Indexer interface {
    EnqueuePageData(ctx context.Context, data models.PageData) error
    StartProcessing(ctx context.Context) error
}

// indexer is an implementation of the Indexer interface.
type indexer struct {
    queue *queue.Queue
	processor processor.Processor
	deduper   processor.Deduper
}

// New creates a new instance of an indexer.
func New() Indexer {
	// Create/initialize the queue.
	pageQueue, err := queue.CreateQueue(1000)
	if err != nil {
		log.Fatalf("failed to create queue: %v", err)
	}
    return &indexer{
        queue:     pageQueue,
        processor: processor.NewProcessor(),
        deduper:   processor.NewDeduper(),
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
                log.Println("context canceled, stopping indexer processing")
                return
            default:
                item, err := i.queue.Remove()
                if err != nil {
                    time.Sleep(200 * time.Millisecond)
                    continue
                }

                cleaned, err := i.processor.CleanAndNormalize(item)
                if err != nil {
                    log.Printf("[SKIP] %s => %v", item.URL, err)
                    continue
                }

                // Now that we have a valid, non-spammy, English doc, let's check duplicates.
                sig := processor.GenerateSignature(cleaned.VisibleText)
                if i.deduper.IsDuplicate(sig) {
                    log.Printf("[SKIP - DUPLICATE] %s => near-duplicate found", cleaned.URL)
                    continue
                }

                // Not a duplicate, so store signature
                i.deduper.StoreSignature(sig)

                // This doc passes all checks. Next step would be NLP, then indexing, etc.
                log.Printf("Document is unique. Ready for next step: %+v\n", cleaned)
            }
        }
    }()
    return nil
}
