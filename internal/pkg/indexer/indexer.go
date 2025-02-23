package indexer

import (
    "bytes"
    "context"
    "encoding/json"
    "indexer/internal/logger"
    "indexer/internal/pkg/models"
    "net/http"
    "sync"
    "go.uber.org/zap"
)

// Buffers documents until a threshold is reached or a flush interval elapses.
type BulkIndexer struct {
    mutex        sync.Mutex
    buffer       []*models.Document
    threshold    int
    flushChannel chan struct{}
    elasticURL string
    indexName  string
}

// Creates a new BulkIndexer.
func NewBulkIndexer(threshold int, elasticURL, indexName string) *BulkIndexer {
    indexer := &BulkIndexer{
        buffer:       make([]*models.Document, 0, threshold),
        threshold:    threshold,
        flushChannel: make(chan struct{}, 1),
        elasticURL:   elasticURL,
        indexName:    indexName,
    }
    go indexer.startFlushing()
    return indexer
}

// Runs in a goroutine and flushes when signaled.
func (indexer *BulkIndexer) startFlushing() {
    for {
        select {
        case <-indexer.flushChannel:
            indexer.flush()
        }
    }
}

// Adds a doc to the buffer and signals flush if threshold is met.
func (indexer *BulkIndexer) AddDocumentToIndexerPayload(doc *models.Document) {
    indexer.mutex.Lock()
    defer indexer.mutex.Unlock()

    indexer.buffer = append(indexer.buffer, doc)
    // If threshold is reached, signal a flush
    if len(indexer.buffer) >= indexer.threshold {
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

    // Build NDJSON
    var ndjsonPayload bytes.Buffer
    for _, doc := range docsToIndex {
        // TODO: define doc IDs, e.g. from doc.URL
        meta := map[string]map[string]string{
            "index": {
                "_index": indexer.indexName,
                // "_id": "custom-id-based-on-url-or-hash",
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

    // Asynchronously send the payload
    go indexer.sendBulkRequest(ndjsonPayload.Bytes())
}

// Sends the bulk payload to Elasticsearch.
func (indexer *BulkIndexer) sendBulkRequest(payload []byte) {
    request, err := http.NewRequestWithContext(context.Background(), "POST", indexer.elasticURL, bytes.NewReader(payload))
    if err != nil {
        logger.Log.Error("Failed to create bulk request", zap.Error(err))
        return
    }
    request.Header.Set("Content-Type", "application/x-ndjson")

    response, err := http.DefaultClient.Do(request)
    if err != nil {
        logger.Log.Error("Bulk request failed", zap.Error(err))
        return
    }
    defer response.Body.Close()

    if response.StatusCode >= 200 && response.StatusCode < 300 {
        logger.Log.Info("Bulk indexing successful", zap.Int("status_code", response.StatusCode))
    } else {
        logger.Log.Warn("Bulk indexing failed", zap.Int("status_code", response.StatusCode))
    }
}
