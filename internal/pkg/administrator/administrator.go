package administrator

import (
    "context"
	"log"
	"time"
	"indexer/internal/pkg/indexer"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/queue"
	"indexer/internal/pkg/processor"
)

const (
	indexThreshold = 3 // Number of documents to buffer before indexing. We will increase this.
)

// Defines the methods that an Administrator should implement.
type Administrator interface {
    EnqueuePageData(ctx context.Context, data models.PageData) error
    ProcessAndIndex(ctx context.Context) error
}

// Implementation of the Administrator interface.
type administrator struct {
	indexer   *indexer.BulkIndexer
    queue     *queue.Queue
	processor processor.Processor
	deduper   processor.Deduper
}

// New creates a new instance of an Administrator.
func New() Administrator {
	// Create/initialize the queue.
	pageQueue, err := queue.CreateQueue(1000)
	if err != nil {
		log.Fatalf("failed to create queue: %v", err)
	}
    return &administrator{
		indexer:   indexer.NewBulkIndexer(indexThreshold, "http://localhost:9200/index"),
        queue:     pageQueue,
        processor: processor.NewProcessor(),
        deduper:   processor.NewDeduper(),

    }
}

func (admin *administrator) EnqueuePageData(ctx context.Context, data models.PageData) error {
    // This quickly returns so the crawler can move on.
    return admin.queue.Insert(data)
}

// 
func (admin *administrator) ProcessAndIndex(ctx context.Context) error {
    go func() {
        for {
            select {
            case <-ctx.Done():
                log.Println("context canceled, stopping indexer processing")
                return
            default:
                pageData, err := admin.queue.Remove()
                if err != nil {
                    time.Sleep(200 * time.Millisecond)
                    continue
                }

				document := models.Document{}

                err = admin.processor.Process(&pageData, &document)
				if err != nil {
					log.Printf("failed to process page: %v", err)
				} else {
					log.Printf("processed page %q", pageData.URL)
				}

				// Send the document to the indexer.
				admin.indexer.AddDocumentToIndexerPayload(&document);
            }
        }
    }()
    return nil
}