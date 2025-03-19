package processor

import (
    "bytes"
    "context"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "sync"
    "time"
    "go.uber.org/zap"
    "golang.org/x/time/rate"
    "indexer/internal/pkg/circuitbreaker"
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/metrics"
)

// Handles NLP processing in batches
type BatchProcessor struct {
    nlpServiceURL  string
    circuitBreaker *circuitbreaker.CircuitBreaker
    batchSize      int
    batchTimeout   time.Duration
    
    // Rate limiter for controlling API request rate
    rateLimiter    *rate.Limiter
    limiterMu      sync.Mutex
    
    // Batch state
    mu             sync.Mutex
    currentBatch   []batchItem
    processingChan chan struct{}
    
    // For graceful shutdown
    done           chan struct{}
}

// Represents a document in the batch
type batchItem struct {
    text         string
	needsSummary bool
    resultCh     chan nlpResult
    timestamp    time.Time
}

// Holds the NLP processing results
type nlpResult struct {
    entities   []entity
    keyphrases []string
    summary    string
    err        error
}

type entity struct {
    Text  string `json:"text"`
    Label string `json:"label"`
}

// Creates a new NLP batch processor
func NewBatchProcessor(nlpServiceURL string, batchSize int, batchTimeout time.Duration) *BatchProcessor {
    bp := &BatchProcessor{
        nlpServiceURL:  nlpServiceURL,
        circuitBreaker: circuitbreaker.NewCircuitBreaker("nlp-service", 5, 30*time.Second),
        batchSize:      batchSize,
        batchTimeout:   batchTimeout,
        // Rate limit to 5 batch requests per second with a burst of 10
        rateLimiter:    rate.NewLimiter(rate.Limit(5), 10),
        currentBatch:   make([]batchItem, 0, batchSize),
        processingChan: make(chan struct{}, 1),
        done:           make(chan struct{}),
    }
    
    // Start batch processing goroutine
    go bp.processBatches()
    
    return bp
}

// Gracefully shuts down the batch processor
func (bp *BatchProcessor) Stop() {
    close(bp.done)
}

// Submits text for NLP processing and returns results
func (bp *BatchProcessor) Process(ctx context.Context, text string) ([]entity, []string, error) {
    
	if text == "" {
        return nil, nil, nil
    }
    
    resultCh := make(chan nlpResult, 1)
    item := batchItem{
        text:         text,
        resultCh:     resultCh,
        timestamp:    time.Now(),
    }
    
    // Add to batch
    bp.mu.Lock()
    bp.currentBatch = append(bp.currentBatch, item)
    
    // If batch is full, trigger processing
    if len(bp.currentBatch) >= bp.batchSize {
        select {
        case bp.processingChan <- struct{}{}:
            // Signal sent successfully
        default:
            // Channel already has signal
        }
    }
    bp.mu.Unlock()
    
    // Wait for result or context cancellation
    select {
    case result := <-resultCh:
        if result.err != nil {
            return nil, nil, result.err
        }
        return result.entities, result.keyphrases, nil
    case <-ctx.Done():
        return nil, nil, ctx.Err()
    }
}

// Runs a loop to process batches when triggered
func (bp *BatchProcessor) processBatches() {
    ticker := time.NewTicker(bp.batchTimeout)
    defer ticker.Stop()
    
    for {
        select {
        case <-bp.done:
            return
        case <-bp.processingChan:
            bp.processBatch()
        case <-ticker.C:
            bp.processBatch()
        }
    }
}

// Handles processing of the current batch
func (bp *BatchProcessor) processBatch() {
    bp.mu.Lock()
    if len(bp.currentBatch) == 0 {
        bp.mu.Unlock()
        return
    }
    
    // Get current batch and reset
    batch := bp.currentBatch
    bp.currentBatch = make([]batchItem, 0, bp.batchSize)
    bp.mu.Unlock()
    
    // Track metrics
    metrics.NlpBatchCount.Inc()
    metrics.NlpBatchSize.Observe(float64(len(batch)))
    
    // Check circuit breaker state
    if bp.circuitBreaker.State() == "open" {
        logger.Log.Warn("Circuit breaker open, skipping NLP batch")
        
        // Return circuit open error to all items
        for _, item := range batch {
            item.resultCh <- nlpResult{
                err: circuitbreaker.ErrCircuitOpen,
            }
        }
        return
    }
    
    // Apply rate limiting before sending the batch
    bp.limiterMu.Lock()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    err := bp.rateLimiter.Wait(ctx)
    cancel()
    bp.limiterMu.Unlock()
    
    if err != nil {
        logger.Log.Warn("Rate limit exceeded for NLP batch", zap.Error(err))
        // Return rate limit error to all items
        for _, item := range batch {
            item.resultCh <- nlpResult{
                err: fmt.Errorf("rate limit exceeded: %w", err),
            }
        }
        return
    }
    
    // Prepare batch request
    documents := make([]map[string]interface{}, len(batch))
    for i, item := range batch {
        documents[i] = map[string]interface{}{
            "text":          item.text,
            "needs_summary": item.needsSummary,
        }
    }
    
    payload := map[string]interface{}{
        "documents": documents,
    }
    
    jsonData, err := json.Marshal(payload)
    if err != nil {
        logger.Log.Error("Failed to marshal NLP batch request", zap.Error(err))
        for _, item := range batch {
            item.resultCh <- nlpResult{err: err}
        }
        return
    }
    
    // Process batch with circuit breaker
    var results map[string]interface{}
    err = bp.circuitBreaker.Execute(func() error {
        start := time.Now()
        
        // Create request with increased timeout for batch
        req, err := http.NewRequest("POST", bp.nlpServiceURL+"/batch", bytes.NewBuffer(jsonData))
        if err != nil {
            return err
        }
        req.Header.Set("Content-Type", "application/json")
        
        // Use longer timeout for batch requests
        client := &http.Client{Timeout: 30 * time.Second}
        resp, err := client.Do(req)
        if err != nil {
            metrics.NlpErrors.Inc()
            return err
        }
        defer resp.Body.Close()
        
        // Track latency
        metrics.NlpLatency.Observe(time.Since(start).Seconds())
        
        if resp.StatusCode != http.StatusOK {
            metrics.NlpErrors.Inc()
            return fmt.Errorf("NLP service returned status: %d", resp.StatusCode)
        }
        
        return json.NewDecoder(resp.Body).Decode(&results)
    })
    
    // Handle circuit breaker error
    if errors.Is(err, circuitbreaker.ErrCircuitOpen) {
        for _, item := range batch {
            item.resultCh <- nlpResult{err: err}
        }
        return
    }
    
    // Handle general error
    if err != nil {
        logger.Log.Error("NLP batch request failed", zap.Error(err))
        for _, item := range batch {
            item.resultCh <- nlpResult{err: err}
        }
        return
    }
    
    // Process results
    resultsList, ok := results["results"].([]interface{})
    if !ok || len(resultsList) != len(batch) {
        err := fmt.Errorf("invalid response format or mismatch in result count")
        logger.Log.Error("NLP batch response error", zap.Error(err))
        for _, item := range batch {
            item.resultCh <- nlpResult{err: err}
        }
        return
    }
    
    // Parse and return results
    for i, rawResult := range resultsList {
        if i >= len(batch) {
            break
        }
        
        result, ok := rawResult.(map[string]interface{})
        if !ok {
            batch[i].resultCh <- nlpResult{err: fmt.Errorf("invalid result format")}
            continue
        }
        
        // Parse entities
        var entities []entity
        if entitiesRaw, ok := result["entities"].([]interface{}); ok {
            for _, e := range entitiesRaw {
                entMap, ok := e.(map[string]interface{})
                if !ok {
                    continue
                }
                
                text, _ := entMap["text"].(string)
                label, _ := entMap["label"].(string)
                
                entities = append(entities, entity{
                    Text:  text,
                    Label: label,
                })
            }
        }
        
        // Parse keyphrases
        var keyphrases []string
        if phrasesRaw, ok := result["keyphrases"].([]interface{}); ok {
            for _, k := range phrasesRaw {
                if kp, ok := k.(string); ok {
                    keyphrases = append(keyphrases, kp)
                }
            }
        }
        
        // Parse summary
        summary, _ := result["summary"].(string)
        
        // Send result back
        batch[i].resultCh <- nlpResult{
            entities:   entities,
            keyphrases: keyphrases,
            summary:    summary,
        }
    }
}