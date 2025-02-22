// cmd/indexer/main.go
package main

import (
    "log"
	"time"
    "github.com/elastic/go-elasticsearch/v8"
    "github.com/elastic/go-elasticsearch/v8/esutil"
    "indexer/internal/pkg/indexer"
    "indexer/internal/pkg/queue"
)

func main() {
	q, err := queue.CreateQueue(1000)
	if err != nil {
		log.Fatalf("Error creating queue: %v", err)
	}

	esClient, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{"http://localhost:9200"},
	})
	if err != nil {
		log.Fatalf("Error creating Elasticsearch client: %v", err)
	}

	bulkIndexer, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:     esClient,
		Index:      "webpages",
		NumWorkers: 4,
		FlushBytes: 5 * 1024 * 1024,
	})
	if err != nil {
		log.Fatalf("Error creating BulkIndexer: %v", err)
	}

	// Instantiate and start the indexer with a 100ms poll interval.
	idx := indexer.New(q, bulkIndexer, 100 * time.Millisecond)
	idx.Start()

	// Your HTTP server or other components can now insert items into the queue.
	// For example:
	// q.Insert([]byte(`{"url": "https://example.com", "title": "Example"}`))

	select {} // block forever (or use a proper graceful shutdown mechanism)
}
