package indexer

import (
    "bytes"
    "context"
    "encoding/json"
    "log"
	"time"
    "github.com/elastic/go-elasticsearch/v8/esutil"
    "indexer/internal/pkg/queue"
)

// Indexer is responsible for reading documents from the queue and
// indexing them into Elasticsearch using the BulkIndexer.
type Indexer struct {
	Queue        *queue.Queue
	BulkIndexer  esutil.BulkIndexer
	PollInterval time.Duration // time to wait when the queue is empty
}

// New creates a new Indexer.
func New(q *queue.Queue, bulkIndexer esutil.BulkIndexer, pollInterval time.Duration) *Indexer {
	return &Indexer{
		Queue:        q,
		BulkIndexer:  bulkIndexer,
		PollInterval: pollInterval,
	}
}

// Start launches a background goroutine that continuously polls the queue
// and indexes any available documents.
func (i *Indexer) Start() {
	go i.processQueue()
}

// processQueue continuously polls the queue for new documents.
func (i *Indexer) processQueue() {
	for {
		// Check if the queue is empty.
		if i.Queue.IsEmpty() {
			// Sleep for the poll interval before checking again.
			time.Sleep(i.PollInterval)
			continue
		}

		// Remove an item from the queue.
		item := i.Queue.Remove()
		if item == nil {
			// Should not happen since IsEmpty() returned false,
			// but safeguard against nil items.
			continue
		}

		// Unmarshal the JSON payload.
		var doc map[string]interface{}
		if err := json.Unmarshal(item, &doc); err != nil {
			log.Printf("Error unmarshaling document: %v", err)
			continue
		}

		// Optionally use a field (e.g. URL) as the document ID.
		docID := ""
		if url, ok := doc["url"].(string); ok {
			docID = url
		}

		// Re-marshal the document to ensure it is in proper JSON format.
		docBytes, err := json.Marshal(doc)
		if err != nil {
			log.Printf("Error marshaling document: %v", err)
			continue
		}

		// Wrap the document in a bytes.Reader, which implements io.ReadSeeker.
		reader := bytes.NewReader(docBytes)

		// Add the document to the BulkIndexer.
		err = i.BulkIndexer.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				Action:     "index", // "create" can be used to avoid overwriting
				DocumentID: docID,
				Body:       reader,
				OnFailure: 	func(
					ctx context.Context,
					item esutil.BulkIndexerItem,
					resp esutil.BulkIndexerResponseItem,
					err error,
				) {
					if err != nil {
						log.Printf("Bulk indexer error: %v", err)
					} else {
						log.Printf("Bulk indexer failure: %s: %s", resp.Error.Type, resp.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			log.Printf("Error adding document to BulkIndexer: %v", err)
		}
	}
}
