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

// Language detection metrics
var (
    // NonEnglishPagesSkipped counts skipped non-English pages
    NonEnglishPagesSkipped = promauto.NewCounter(prometheus.CounterOpts{
        Name: "indexer_non_english_pages_skipped_total",
        Help: "Total number of pages skipped because they were not in English",
    })

    // LanguageDetectionFailures counts language detection failures
    LanguageDetectionFailures = promauto.NewCounter(prometheus.CounterOpts{
        Name: "indexer_language_detection_failures_total",
        Help: "Total number of language detection failures",
    })

    // LanguageDetectionLatency measures time taken for language detection
    LanguageDetectionLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "indexer_language_detection_latency_seconds",
        Help: "Time taken to detect language",
        Buckets: prometheus.DefBuckets,
    })
)

// Spam detection metrics
var (
    HighSpamPagesSkipped = promauto.NewCounter(prometheus.CounterOpts{
        Name: "indexer_high_spam_pages_skipped_total",
        Help: "Total number of pages skipped due to high spam score",
    })
    
    SpamScoreHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "indexer_spam_score_distribution",
        Help: "Distribution of spam scores for processed pages",
        Buckets: []float64{0, 1, 2, 5, 10, 15, 20, 30, 50, 100},
    })
    
    SpamDetectionLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name: "indexer_spam_detection_latency_seconds",
        Help: "Time taken to perform spam detection",
        Buckets: prometheus.DefBuckets,
    })
)

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
