package administrator

import (
	"context"
	"log"
	"time"
	"go.uber.org/zap"
    "indexer/internal/pkg/config"
	"indexer/internal/pkg/logger"
	"indexer/internal/pkg/deduplicator"
	"indexer/internal/pkg/indexer"
	"indexer/internal/pkg/models"
	"indexer/internal/pkg/processor"
	"indexer/internal/pkg/queue"
)

// Administrator interface
type Administrator interface {
    EnqueuePageData(ctx context.Context, data models.PageData) error
    ProcessAndIndex(ctx context.Context) error
    StartService(port string)
    Stop()
}

// Implementation of the Administrator interface
type administrator struct {
    indexer   *indexer.BulkIndexer
    queue     *queue.Queue
    processor processor.Processor
}

// Creates a new instance of an Administrator with a config
func New(config *config.Config) Administrator {
    pageQueue, err := queue.CreateQueue(config.QueueCapacity)
    if err != nil {
        logger.Log.Fatal("Failed to create queue", zap.Error(err))
    }

    // Requires a redis instance. docker run -p 6379:6379 --name redis -d redis:6.2
    deduper, err := deduper.NewRedisDeduper(config)
    if err != nil {
        log.Fatalf("Failed to create deduper: %v", err)
    }

    bulkIndexer := indexer.NewBulkIndexer(
        config.BulkThreshold,
        config.ElasticsearchURL,
        config.IndexName,
        config.FlushInterval,
        config.MaxRetries,
    )

    return &administrator{
        indexer:   bulkIndexer,
        queue:     pageQueue,
        processor: processor.NewProcessor(deduper, config.NlpServiceURL),
    }
}

func (admin *administrator) EnqueuePageData(ctx context.Context, data models.PageData) error {
    // This quickly returns so the crawler can move on
    return admin.queue.Insert(data)
}

// Processes and indexes the page data
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
                    time.Sleep(200 * time.Millisecond)
                    continue
                }

                var document models.Document
                err = admin.processor.Process(&pageData, &document)
                if err != nil {
                    logger.Log.Warn("Failed to process page",
                        zap.String("url", pageData.URL),
                        zap.Error(err))
                } else {
                    logger.Log.Debug("Processed page", zap.String("url", pageData.URL))
                }
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

// Stops the BulkIndexer gracefully
func (admin *administrator) Stop() {
    admin.indexer.Stop()
}
