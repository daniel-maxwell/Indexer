package processor

import (
	"bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "errors"
    "sync"
    "context"
    "golang.org/x/time/rate"
    "go.uber.org/zap"
	"indexer/internal/pkg/models"
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/circuitbreaker"
)

// Defines the interface for adding additional metadata to a document.
type Enricher interface {
	Enrich(pageData *models.PageData, doc *models.Document) error
}

// Implementation of Enricher.
type nlpEnricher struct {
    nlpServiceURL  string
    circuitBreaker *circuitbreaker.CircuitBreaker
    rateLimiter    *rate.Limiter
    limiterMu      sync.Mutex
}

// Creates a new instance of an NLP-based Enricher.
func NewNLPEnricher(nlpServiceURL string) Enricher {
    return &nlpEnricher{
        nlpServiceURL:  nlpServiceURL,
        circuitBreaker: circuitbreaker.NewCircuitBreaker("nlp-service", 5, 30*time.Second),
        // Creates a rate limiter that allows 20 requests per second with a burst of 30
        // These numbers should be tuned based on NLP service capacity
        rateLimiter:    rate.NewLimiter(rate.Limit(20), 30),
    }
}

// Augments the document with entities, keywords, and a summary.
func (enricher *nlpEnricher) Enrich(pageData *models.PageData, doc *models.Document) error {
    if pageData.VisibleText == "" {
        return nil
    }

    // Rate limit before making request
    enricher.limiterMu.Lock()
    if err := enricher.rateLimiter.Wait(context.Background()); err != nil {
        enricher.limiterMu.Unlock()
        logger.Log.Warn("Rate limit wait error", zap.Error(err))
        return nil
    }
    enricher.limiterMu.Unlock()
    
    payload := map[string]string{"text": pageData.VisibleText}
    jsonBody, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal JSON for NLP request: %w", err)
    }
    
    var result struct {
        Entities []struct {
            Text  string `json:"text"`
            Label string `json:"label"`
        } `json:"entities"`
        Keyphrases []string `json:"keyphrases"`
        Summary    string   `json:"summary"`
    }
    
    // Use circuit breaker to execute request with retries
    err = enricher.circuitBreaker.Execute(func() error {
        request, err := http.NewRequest("POST", enricher.nlpServiceURL, bytes.NewBuffer(jsonBody))
        if err != nil {
            return fmt.Errorf("failed to create request: %w", err)
        }
        request.Header.Set("Content-Type", "application/json")
        
        // Timeout to accommodate potential NLP service load
        client := &http.Client{Timeout: 10 * time.Second}
        response, err := client.Do(request)
        if err != nil {
            return fmt.Errorf("NLP service call failed: %w", err)
        }
        defer response.Body.Close()
        
        if response.StatusCode != http.StatusOK {
            return fmt.Errorf("NLP service returned status %d", response.StatusCode)
        }
        
        if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
            return fmt.Errorf("failed to decode response: %w", err)
        }
        
        return nil
    })
    
    // If circuit is open, just return and continue without NLP enrichment
    if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
        logger.Log.Warn("Skipping NLP enrichment, circuit breaker open")
        return nil
    }
    
    // For other errors, log but continue
    if err != nil {
        logger.Log.Warn("NLP enrichment failed", zap.Error(err))
        return nil
    }
    
    // Process results and update document
    var entities []string
    for _, ent := range result.Entities {
        entities = append(entities, fmt.Sprintf("%s: %s", ent.Label, ent.Text))
    }
    doc.Entities = entities
    doc.Keywords = result.Keyphrases
    doc.Summary = result.Summary
    
    return nil
}
