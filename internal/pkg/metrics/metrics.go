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

// NLP service metrics
var (
    NlpRequests = promauto.NewCounter(prometheus.CounterOpts{
        Name: "indexer_nlp_requests_total",
        Help: "Total number of requests sent to the NLP service",
    })
    
    NlpErrors = promauto.NewCounter(prometheus.CounterOpts{
        Name: "indexer_nlp_errors_total",
        Help: "Total number of failed requests to the NLP service",
    })
    
    NlpLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "indexer_nlp_latency_seconds",
        Help: "Time taken to process NLP requests",
        Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // From 100ms to ~100s
    })
    
    NlpBatchCount = promauto.NewCounter(prometheus.CounterOpts{
        Name: "indexer_nlp_batch_count_total",
        Help: "Total number of batches sent to the NLP service",
    })
    
    NlpBatchSize = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "indexer_nlp_batch_size",
        Help: "Size of batches sent to the NLP service",
        Buckets: []float64{1, 2, 5, 10, 20, 50, 100},
    })
    
    CircuitBreakerState = promauto.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "indexer_circuit_breaker_state",
            Help: "Current state of circuit breakers (0=closed, 1=half-open, 2=open)",
        },
        []string{"service"},
    )
)
