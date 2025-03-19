package processor

import (
    "context"
    "fmt"
    "time"
    "go.uber.org/zap"
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/metrics"
    "indexer/internal/pkg/models"
)

// Defines the interface for adding additional metadata to a document.
type Enricher interface {
    Enrich(pageData *models.PageData, doc *models.Document) error
}

// Implementation of Enricher.
type nlpEnricher struct {
    batchProcessor *BatchProcessor
}

// Creates a new instance of an NLP-based Enricher.
func NewNLPEnricher(nlpServiceURL string) Enricher {
    // Default batch settings for now
    batchSize := 10  // Process 10 documents at a time
    batchTimeout := 200 * time.Millisecond
    return &nlpEnricher{
        batchProcessor: NewBatchProcessor(nlpServiceURL, batchSize, batchTimeout),
    }
}

// Augments the document with entities and keywords using batch processing.
func (enricher *nlpEnricher) Enrich(pageData *models.PageData, doc *models.Document) error {
    // Skip if no text
    if pageData.VisibleText == "" {
        return nil
    }
    
    // Create context with timeout for processing
    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()
    
    // Record timing for metrics
    startTime := time.Now()
    
    // Process through batch processor
    entities, keyphrases, err := enricher.batchProcessor.Process(ctx, pageData.VisibleText)
    
    // Update metrics
    metrics.NlpRequests.Inc()
    metrics.NlpLatency.Observe(time.Since(startTime).Seconds())
    
    if err != nil {
        logger.Log.Warn("NLP enrichment failed", zap.Error(err), zap.String("url", pageData.URL))
        metrics.NlpErrors.Inc()
        // Continue without NLP enrichment
        return nil
    }
    
    // Map entities to doc.Entities
    var docEntities []string
    for _, ent := range entities {
        docEntities = append(docEntities, fmt.Sprintf("%s: %s", ent.Label, ent.Text))
    }
    doc.Entities = docEntities
    
    // Store keywords
    doc.Keywords = keyphrases
    
    // Copy basic fields from PageData to Document
    doc.URL = pageData.URL
    doc.CanonicalURL = pageData.CanonicalURL
    doc.Title = pageData.Title
    doc.MetaDescription = pageData.MetaDescription
    doc.VisibleText = pageData.VisibleText
    doc.InternalLinks = pageData.InternalLinks
    doc.ExternalLinks = pageData.ExternalLinks
    doc.DatePublished = pageData.DatePublished
    doc.DateModified = pageData.DateModified
    doc.SocialLinks = pageData.SocialLinks
    doc.IsSecure = pageData.IsSecure
    
    if pageData.LoadTime > 0 {
        doc.LoadTime = int64(pageData.LoadTime / time.Millisecond)
    }
    
    // Set last crawled time
    doc.LastCrawled = time.Now()
    
    return nil
}