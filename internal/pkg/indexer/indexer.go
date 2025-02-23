package indexer

import (
	"bytes"
	"context"
	"encoding/json"
	"indexer/internal/pkg/models"
	"log"
	"net/http"
	"sync"
)

// Buffers documents until a threshold is reached or a flush interval elapses.
type BulkIndexer struct {
	mutex         sync.Mutex
	buffer        []*models.Document
	threshold     int
	flushChannel  chan struct{}
	elasticURL    string // elasticURL is the endpoint for the Elasticsearch Bulk API.
}

// Creates a new BulkIndexer.
func NewBulkIndexer(threshold int, elasticURL string) *BulkIndexer {
	bulkIndexer := &BulkIndexer{
		buffer:        make([]*models.Document, 0, threshold),
		threshold:     threshold,
		flushChannel:  make(chan struct{}, 1),
		elasticURL:    elasticURL,
	}
	go bulkIndexer.startFlushing()
	return bulkIndexer
}

// Runs in a goroutine and flushes the buffer periodically or when signaled.
func (bulkIndexer *BulkIndexer) startFlushing() {
	for {
		select {
		case <-bulkIndexer.flushChannel:
			bulkIndexer.flush()
		}
	}
}

// Adds a document to the buffer.
// If the threshold is reached, it signals a flush.
func (bulkIndexer *BulkIndexer) AddDocumentToIndexerPayload(doc *models.Document) {
	bulkIndexer.mutex.Lock()
	bulkIndexer.buffer = append(bulkIndexer.buffer, doc)
	// Signal flush if threshold is reached.
	if len(bulkIndexer.buffer) >= bulkIndexer.threshold {
		select {
		case bulkIndexer.flushChannel <- struct{}{}:
		default:
			// flush already signaled
		}
	}
	bulkIndexer.mutex.Unlock()
}

// Builds an NDJSON payload from buffered documents and sends it to Elasticsearch.
func (bulkIndexer *BulkIndexer) flush() {
	bulkIndexer.mutex.Lock()
	if len(bulkIndexer.buffer) == 0 {
		bulkIndexer.mutex.Unlock()
		return
	}
	docsToIndex := bulkIndexer.buffer
	bulkIndexer.buffer = make([]*models.Document, 0, bulkIndexer.threshold)
	bulkIndexer.mutex.Unlock()

	var ndjsonPayload bytes.Buffer
	for _, doc := range docsToIndex {
		// Build the metadata action line for indexing.
		meta := map[string]map[string]string{
			"index": {
				"_index": "documents", // Adjust index name as needed.
				// Optionally include "_id": "your-document-id"
			},
		}
		metaLine, err := json.Marshal(meta)
		if err != nil {
			log.Printf("failed to marshal meta line: %v", err)
			continue
		}
		ndjsonPayload.Write(metaLine)
		ndjsonPayload.WriteByte('\n')

		// Marshal the actual document.
		docLine, err := json.Marshal(doc)
		if err != nil {
			log.Printf("failed to marshal document: %v", err)
			continue
		}
		ndjsonPayload.Write(docLine)
		ndjsonPayload.WriteByte('\n')
	}

	log.Printf("Flushing %d documents to Elasticsearch", len(docsToIndex))
	// For now, we log the NDJSON payload.
	log.Printf("NDJSON payload:\n%s", ndjsonPayload.String())

	// Asynchronously send the payload to Elasticsearch.
	go bulkIndexer.sendBulkRequest(ndjsonPayload.Bytes())
}

// Sends the NDJSON payload to Elasticsearch's Bulk API.
func (bulkIndexer *BulkIndexer) sendBulkRequest(payload []byte) {
	request, err := http.NewRequestWithContext(context.Background(), "POST", bulkIndexer.elasticURL, bytes.NewReader(payload))
	if err != nil {
		log.Printf("failed to create bulk request: %v", err)
		return
	}
	request.Header.Set("Content-Type", "application/x-ndjson")

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		log.Printf("bulk request failed: %v", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		log.Println("bulk indexing successful")
	} else {
		log.Printf("bulk indexing failed with status: %s", response.Status)
	}
}