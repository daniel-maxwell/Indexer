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
    doc.Language = pageData.Language
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

    doc.QualityScore = enricher.calculateQualityScore(doc)
    
    // Set last crawled time
    doc.LastCrawled = time.Now()
    
    return nil
}

// Quality scoring for prioritization
func (enricher *nlpEnricher) calculateQualityScore(doc *models.Document) int {
    score := 0
    
    // Text quality factors
    if len(doc.VisibleText) > 100 {
        score += 10
    }
    if len(doc.Title) > 5 && len(doc.Title) < 150 {
        score += 10
    }
    if len(doc.MetaDescription) > 50 {
        score += 5
    }
    
    // Content signals
    if len(doc.Entities) >= 1 {
        score += 10
    }
    if len(doc.Keywords) > 3 {
        score += 10
    }
    
    // Link signals
    if len(doc.InternalLinks) > 0 {
        score += 5
    }
    if len(doc.ExternalLinks) > 0 {
        score += 5
    }

    if doc.Language == "en" {
        score += 10
    }
    
    // Technical signals
    if doc.IsSecure {
        score += 25
    }
    
    if doc.LoadTime < 1000 {  // Less than 1 second
        score += 10
    } else if doc.LoadTime < 2000 {  // Less than 2 seconds
        score += 5
    } else if doc.LoadTime < 3000 {
        score += 2
    }
    
    // Cap at 100
    if score > 100 {
        score = 100
    }
    
    return score
}