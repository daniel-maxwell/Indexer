package administrator

import (
    "context"
    "indexer/internal/config"
    "indexer/internal/logger"
    "indexer/internal/pkg/indexer"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/processor"
    "indexer/internal/pkg/queue"
    "go.uber.org/zap"
    "time"
)

// Administrator interface remains the same
type Administrator interface {
    EnqueuePageData(ctx context.Context, data models.PageData) error
    ProcessAndIndex(ctx context.Context) error
    StartService(port string) // updated signature to accept port
}

type administrator struct {
    indexer   *indexer.BulkIndexer
    queue     *queue.Queue
    processor processor.Processor
}

// New creates a new instance of an Administrator with a config
func New(cfg *config.Config) Administrator {
    // Create/initialize the queue
    pageQueue, err := queue.CreateQueue(cfg.QueueCapacity)
    if err != nil {
        logger.Log.Fatal("Failed to create queue", zap.Error(err))
    }

    deduper := processor.NewDeduper()

    // Initialize the BulkIndexer with config
    bulkIndexer := indexer.NewBulkIndexer(cfg.BulkThreshold, cfg.ElasticsearchURL, cfg.IndexName)

    return &administrator{
        indexer:   bulkIndexer,
        queue:     pageQueue,
        processor: processor.NewProcessor(deduper),
    }
}

func (admin *administrator) EnqueuePageData(ctx context.Context, data models.PageData) error {
    // This quickly returns so the crawler can move on
    return admin.queue.Insert(data)
}

func (admin *administrator) ProcessAndIndex(ctx context.Context) error {
    go func() {
        for {
            select {
            case <-ctx.Done():
                logger.Log.Info("Context canceled, stopping indexer processing")
                return
            default:
                pageData, err := admin.queue.Remove()
                if err != nil {
                    // queue is empty, wait briefly
                    time.Sleep(200 * time.Millisecond)
                    continue
                }

                document := models.Document{}
                err = admin.processor.Process(&pageData, &document)
                if err != nil {
                    logger.Log.Warn("Failed to process page", zap.String("url", pageData.URL), zap.Error(err))
                } else {
                    logger.Log.Debug("Processed page", zap.String("url", pageData.URL))
                }

                // Send the document to the indexer
                admin.indexer.AddDocumentToIndexerPayload(&document)
            }
        }
    }()
    return nil
}

// StartService starts the HTTP ingest service at the given port
func (admin *administrator) StartService(port string) {
    logger.Log.Info("Starting HTTP ingestion service", zap.String("port", port))
    // The code is in ingest_service.go, we pass the port in
    startIngestHTTP(admin, port)
}
