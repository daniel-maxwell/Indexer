// internal/pkg/administrator/administrator.go (modified)
package administrator

import (
    "context"
    "time"
    "go.uber.org/zap"
    "indexer/internal/pkg/config"
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/deduplicator"
    "indexer/internal/pkg/indexer"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/processor"
    "indexer/internal/pkg/queue"
    "indexer/internal/pkg/worker"
)

// Administrator interface
type Administrator interface {
    EnqueuePageData(ctx context.Context, data models.PageData) error
    ProcessAndIndex(ctx context.Context) error
    StartService(port string)
    Stop()
    QueueDepth() int
    WorkerCount() int
    StartTime() time.Time
}

// Implementation of the Administrator interface
type administrator struct {
    indexer     *indexer.BulkIndexer
    queue       *queue.Queue
    processor   processor.Processor
    workerPool  *worker.WorkerPool
    startTime   time.Time
    numWorkers  int
}

// Creates a new instance of an Administrator with a config
func New(config *config.Config) Administrator {
    pageQueue, err := queue.CreateQueue(config.QueueCapacity)
    if err != nil {
        logger.Log.Fatal("Failed to create queue", zap.Error(err))
    }

    deduper, err := deduper.NewRedisDeduper(config)
    if err != nil {
        logger.Log.Fatal("Failed to create deduper", zap.Error(err))
    }

    bulkIndexer := indexer.NewBulkIndexer(
        config.BulkThreshold,
        config.ElasticsearchURL,
        config.IndexName,
        config.FlushInterval,
        config.MaxRetries,
    )

    proc := processor.NewProcessor(deduper, config.NlpServiceURL, config.SpamBlockThreshold)
    
    // Get number of workers from config
    numWorkers := config.NumWorkers
    if numWorkers <= 0 {
        numWorkers = 1 // Default to 1 worker if not specified
    }
    
    wp := worker.NewWorkerPool(numWorkers, pageQueue, proc, bulkIndexer)
    
    return &administrator{
        indexer:     bulkIndexer,
        queue:       pageQueue,
        processor:   proc,
        workerPool:  wp,
        startTime:   time.Now(),
        numWorkers:  numWorkers,
    }
}

func (admin *administrator) EnqueuePageData(ctx context.Context, data models.PageData) error {
    // This quickly returns so the crawler can move on
    return admin.queue.Insert(data)
}

// Processes and indexes the page data with parallel workers
func (admin *administrator) ProcessAndIndex(ctx context.Context) error {
    // Start the worker pool with the provided context
    admin.workerPool.Start(ctx)
    return nil
}

// StartService starts the HTTP ingest service at the given port
func (admin *administrator) StartService(port string) {
    logger.Log.Info("Starting HTTP ingestion service", zap.String("port", port))
    startIngestHTTP(admin, port)
}

// Stops the BulkIndexer and worker pool gracefully
func (admin *administrator) Stop() {
    logger.Log.Info("Beginning shutdown sequence")
    
    // First flush and stop accepting new items in the queue
    admin.queue.Close() // Assuming queue has a Close method to stop accepting new items
    
    logger.Log.Info("Waiting for worker pool to finish processing existing items")
    // Wait for workers to finish current work
    admin.workerPool.Wait()
    
    logger.Log.Info("Worker pool shutdown complete, stopping bulk indexer")
    // Then stop the BulkIndexer and wait for pending requests
    admin.indexer.Stop()
    
    logger.Log.Info("Administrator stopped gracefully")
}

// Returns the current queue depth for health checks
func (admin *administrator) QueueDepth() int {
    return admin.queue.Length()
}

// Returns the number of workers for health checks
func (admin *administrator) WorkerCount() int {
    return admin.numWorkers
}

// Returns when the service was started for health checks
func (admin *administrator) StartTime() time.Time {
    return admin.startTime
}