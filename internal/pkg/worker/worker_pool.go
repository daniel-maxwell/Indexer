package worker

import (
    "context"
    "sync"
    "time"
    
    "go.uber.org/zap"
    
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/processor"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/queue"
    "indexer/internal/pkg/indexer"
    "indexer/internal/pkg/metrics"
)

// Manages a pool of workers that process queue items in parallel
type WorkerPool struct {
    numWorkers     int
    queue          *queue.Queue
    processor      processor.Processor
    indexer        *indexer.BulkIndexer
    wg             sync.WaitGroup
}

// Creates a new worker pool with the specified number of workers
func NewWorkerPool(numWorkers int, queue *queue.Queue, processor processor.Processor, indexer *indexer.BulkIndexer) *WorkerPool {
    return &WorkerPool{
        numWorkers: numWorkers,
        queue:      queue,
        processor:  processor,
        indexer:    indexer,
    }
}

// Launches the worker goroutines
func (wp *WorkerPool) Start(ctx context.Context) {
    logger.Log.Info("Starting worker pool", zap.Int("workers", wp.numWorkers))
    
    for i := 0; i < wp.numWorkers; i++ {
        wp.wg.Add(1)
        go wp.runWorker(ctx, i)
    }
}

// Blocks until all workers have finished
func (wp *WorkerPool) Wait() {
    wp.wg.Wait()
}

// The main loop for each worker goroutine
func (wp *WorkerPool) runWorker(ctx context.Context, id int) {
    defer wp.wg.Done()
    
    logger.Log.Info("Worker started", zap.Int("worker_id", id))
    
    for {
        select {
        case <-ctx.Done():
            logger.Log.Info("Worker received stop signal", zap.Int("worker_id", id))
            return
        default:
            pageData, err := wp.queue.Remove()
            if err != nil {
                // If queue is empty, wait a bit before trying again
                time.Sleep(200 * time.Millisecond)
                continue
            }
            
            var document models.Document
            err = wp.processor.Process(&pageData, &document)
            if err != nil {
                logger.Log.Warn("Failed to process page",
                    zap.Int("worker_id", id),
                    zap.String("url", pageData.URL),
                    zap.Error(err))
                
                if err.Error() == "duplicate page detected" {
                    metrics.DuplicatesDetected.Inc()
                }
            } else {
                logger.Log.Debug("Processed page", 
                    zap.Int("worker_id", id),
                    zap.String("url", pageData.URL))
                
                // Add the document to the indexer
                wp.indexer.AddDocumentToIndexerPayload(&document)
            }
        }
    }
}