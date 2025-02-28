package processor

import (
	"bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "go.uber.org/zap"
	"indexer/internal/pkg/models"
    "indexer/internal/pkg/logger"
)

// Defines the interface for adding additional metadata to a document.
type Enricher interface {
	Enrich(pageData *models.PageData, doc *models.Document) error
}

// Implementation of Enricher.
type nlpEnricher struct{
	nlpServiceURL string
}

// Creates a new instance of an NLP-based Enricher.
func NewNLPEnricher(nlpServiceURL string) Enricher {
	return &nlpEnricher{
        nlpServiceURL: nlpServiceURL,
    }
}

// Augments the document with entities, keywords, and a summary.
func (enricher *nlpEnricher) Enrich(pageData *models.PageData, doc *models.Document) error {
    // Send pageData.VisibleText to the Python microservice
    if pageData.VisibleText == "" {
        // no text, skip
        return nil
    }

    payload := map[string]string{"text": pageData.VisibleText}
    jsonBody, err := json.Marshal(payload)
    if err != nil {
        return fmt.Errorf("failed to marshal JSON for NLP request: %w", err)
    }

    request, err := http.NewRequest("POST", enricher.nlpServiceURL, bytes.NewBuffer(jsonBody))
    if err != nil {
        return fmt.Errorf("failed to create request for NLP service: %w", err)
    }
    request.Header.Set("Content-Type", "application/json")

    client := &http.Client{Timeout: 5 * time.Second}
    response, err := client.Do(request)
    if err != nil {
        logger.Log.Warn("NLP service call failed", zap.Error(err))
        // Optionally skip or fallback
        return nil
    }
    defer response.Body.Close()

    if response.StatusCode != http.StatusOK {
        logger.Log.Warn("NLP service returned non-200", zap.Int("status", response.StatusCode))
        return nil
    }

    var result struct {
        Entities []struct {
            Text  string `json:"text"`
            Label string `json:"label"`
        } `json:"entities"`
        Keyphrases []string `json:"keyphrases"`
        Summary    string   `json:"summary"`
    }

    if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
        logger.Log.Warn("Failed to decode NLP service response", zap.Error(err))
        return nil
    }

    // Map entities to doc.Entities
    var entities []string
    for _, ent := range result.Entities {
        // e.g. "PERSON: John" or "ORG: Google"
        entities = append(entities, fmt.Sprintf("%s: %s", ent.Label, ent.Text))
    }
    doc.Entities = entities

    // Store keywords
    doc.Keywords = result.Keyphrases

    // Store summary
    doc.Summary = result.Summary

    return nil
}
