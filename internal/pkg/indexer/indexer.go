package indexer

import (
    "bytes"
    "context"
    "encoding/json"
    "math/rand"
    "net/http"
    "strings"
    "sync"
    "time"
    "go.uber.org/zap"
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/models"
    "indexer/internal/pkg/metrics"
)

// Buffers documents until threshold or flush interval is reached.
type BulkIndexer struct {
    mutex         sync.Mutex
    buffer        []*models.Document
    threshold     int
    flushChannel  chan struct{}

    elasticURL    string
    indexName     string

    flushInterval time.Duration
    maxRetries    int
    wg            sync.WaitGroup

    
    done chan struct{} // for stopping the flush goroutine
}

// Creates a new BulkIndexer.
func NewBulkIndexer(threshold int, elasticURL, indexName string, flushIntervalSeconds, maxRetries int) *BulkIndexer {
    indexer := &BulkIndexer{
        buffer:         make([]*models.Document, 0, threshold),
        threshold:      threshold,
        flushChannel:   make(chan struct{}, 1),
        elasticURL:     elasticURL,
        indexName:      indexName,
        flushInterval:  time.Duration(flushIntervalSeconds) * time.Second,
        maxRetries:     maxRetries,
        done:           make(chan struct{}),
    }
    go indexer.startFlushing()
    return indexer
}

// Runs in a goroutine and triggers flush on signal or interval
func (indexer *BulkIndexer) startFlushing() {
    ticker := time.NewTicker(indexer.flushInterval)
    defer ticker.Stop()

    for {
        select {
        case <-indexer.done:
            // Final flush before shutdown
            logger.Log.Info("BulkIndexer received done signal, flushing before exit")
            indexer.flush()
            return
        case <-indexer.flushChannel:
            indexer.flush()
        case <-ticker.C:
            indexer.flush()
        }
    }
}

// Adds a doc to the buffer and signals flush if threshold is met.
func (indexer *BulkIndexer) AddDocumentToIndexerPayload(doc *models.Document) {
    indexer.mutex.Lock()
    indexer.buffer = append(indexer.buffer, doc)
    count := len(indexer.buffer)
    indexer.mutex.Unlock()

    // If threshold is reached, signal a flush
    if count >= indexer.threshold {
        select {
        case indexer.flushChannel <- struct{}{}:
        default:
            // flush already signaled
        }
    }
}

// Builds NDJSON payload and sends it to Elasticsearch.
func (indexer *BulkIndexer) flush() {
    indexer.mutex.Lock()
    if len(indexer.buffer) == 0 {
        indexer.mutex.Unlock()
        return
    }
    docsToIndex := indexer.buffer
    indexer.buffer = make([]*models.Document, 0, indexer.threshold)
    indexer.mutex.Unlock()

    metrics.BulkFlushes.Inc()

    // Build NDJSON
    var ndjsonPayload bytes.Buffer
    for _, doc := range docsToIndex {
        // Generate doc ID from URL or canonical URL
        docID := generateDocID(doc.URL, doc.CanonicalURL)
        meta := map[string]map[string]string{
            "index": {
                "_index": indexer.indexName,
                "_id":    docID,
            },
        }
        metaLine, err := json.Marshal(meta)
        if err != nil {
            logger.Log.Error("Failed to marshal meta line", zap.Error(err))
            continue
        }
        ndjsonPayload.Write(metaLine)
        ndjsonPayload.WriteByte('\n')

        docLine, err := json.Marshal(doc)
        if err != nil {
            logger.Log.Error("Failed to marshal document", zap.Error(err))
            continue
        }
        ndjsonPayload.Write(docLine)
        ndjsonPayload.WriteByte('\n')
    }

    logger.Log.Info("Flushing documents to Elasticsearch", zap.Int("count", len(docsToIndex)))
    indexer.wg.Add(1)
    go func() {
        defer indexer.wg.Done()
        indexer.sendBulkRequest(ndjsonPayload.Bytes(), 0)
    }()
}

// Gracefully stops the BulkIndexer (e.g., called during shutdown).
func (indexer *BulkIndexer) Stop() {
    close(indexer.done)
    indexer.wg.Wait() // Wait for in-flight requests to finish
}

// Tries to POST the NDJSON to Elasticsearch, with optional retries.
func (indexer *BulkIndexer) sendBulkRequest(payload []byte, attempt int) {
    request, err := http.NewRequestWithContext(context.Background(), "POST", indexer.elasticURL, bytes.NewReader(payload))
    if err != nil {
        logger.Log.Error("Failed to create bulk request", zap.Error(err))
        return
    }
    request.Header.Set("Content-Type", "application/x-ndjson")

    response, err := http.DefaultClient.Do(request)
    if err != nil {
        logger.Log.Error("Bulk request failed", zap.Error(err), zap.Int("attempt", attempt))
        // Retry if we haven't exceeded maxRetries
        if attempt < indexer.maxRetries {
            time.Sleep(backoffDuration(attempt))
            indexer.sendBulkRequest(payload, attempt + 1)
        } else {
            metrics.BulkFailures.Inc()
        }
        return
    }
    defer response.Body.Close()

    if response.StatusCode >= 200 && response.StatusCode < 300 {
        logger.Log.Info("Bulk indexing successful", zap.Int("status_code", response.StatusCode))
    } else {
        logger.Log.Warn("Bulk indexing failed", zap.Int("status_code", response.StatusCode), zap.Int("attempt", attempt))
        // Retry on non-2xx if we haven't exceeded maxRetries
        if attempt < indexer.maxRetries {
            time.Sleep(backoffDuration(attempt))
            indexer.sendBulkRequest(payload, attempt+1)
        } else {
            metrics.BulkFailures.Inc()
        }
    }
}

// Returns a simple exponential backoff time.
func backoffDuration(attempt int) time.Duration {
    base := time.Second
    // Exponential backoff, plus a little jitter
    backoff := time.Duration(1<<attempt) * base
    jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
    return backoff + jitter
}

// Returns a stable ID based on canonicalURL if available, else URL.
// Additional hashing or slugification may be used for a consistent ID in future.
func generateDocID(urlStr, canonicalStr string) string {
    if strings.TrimSpace(canonicalStr) != "" {
        return sanitizeID(canonicalStr)
    }
    return sanitizeID(urlStr)
}

// Sanitize the ID to remove problematic characters and ensure it's URL-safe.
func sanitizeID(raw string) string {
    // Remove protocols
    clean := strings.ReplaceAll(raw, "http://", "")
    clean = strings.ReplaceAll(clean, "https://", "")
    
    // Replace problematic characters
    clean = strings.ReplaceAll(clean, "/", "_")
    clean = strings.ReplaceAll(clean, "?", "_")
    clean = strings.ReplaceAll(clean, "&", "_")
    clean = strings.ReplaceAll(clean, "=", "_")
    clean = strings.ReplaceAll(clean, "#", "_")
    clean = strings.ReplaceAll(clean, " ", "_")
    clean = strings.ReplaceAll(clean, ":", "_")
    
    // Remove any remaining invalid characters
    var result strings.Builder
    for _, r := range clean {
        if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '.' || r == '-' {
            result.WriteRune(r)
        }
    }
    
    // Keep it short
    resultStr := result.String()
    if len(resultStr) > 100 {
        resultStr = resultStr[:100]
    }
    
    return resultStr
}
