package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Counts how many pages have been processed in total.
var PagesProcessed = promauto.NewCounter(prometheus.CounterOpts{
    Name: "indexer_pages_processed_total",
    Help: "Total number of pages processed successfully",
})

// Counts how many pages were flagged as duplicates.
var DuplicatesDetected = promauto.NewCounter(prometheus.CounterOpts{
    Name: "indexer_duplicates_detected_total",
    Help: "Total number of pages that were flagged as duplicates",
})

// Measures how many documents have been sent to ES.
var DocumentsIndexed = promauto.NewCounter(prometheus.CounterOpts{
    Name: "indexer_documents_indexed_total",
    Help: "Total number of documents flushed to Elasticsearch",
})

// Captures how many times we performed a bulk flush operation.
var BulkFlushes = promauto.NewCounter(prometheus.CounterOpts{
    Name: "indexer_bulk_flushes_total",
    Help: "Total number of times documents were flushed in bulk to Elasticsearch",
})

// Captures how many times a bulk request failed.
var BulkFailures = promauto.NewCounter(prometheus.CounterOpts{
    Name: "indexer_bulk_failures_total",
    Help: "Total number of bulk requests that failed",
})
