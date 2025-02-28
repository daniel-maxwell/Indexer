package indexer

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
	"go.uber.org/zap"
	"indexer/internal/pkg/models"
	"indexer/internal/pkg/logger"
)

func init() {
	// Ensure that the logger is not nil during tests.
	logger.Log = zap.NewNop()
}

// Verifies that when the threshold is met, the BulkIndexer 
// flushes documents to the (simulated) Elasticsearch endpoint.
func TestBulkIndexerFlushSuccess(t *testing.T) {
	// Create a channel to capture the request payload.
	payloadCh := make(chan []byte, 1)

	// Create a test server that always returns a 200 OK.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("Failed to read request body: %v", err)
		}
		payloadCh <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer testServer.Close()

	// Create a BulkIndexer with a threshold of 2 docs and a long flush interval (so flush is triggered by threshold).
	threshold := 2
	flushIntervalSeconds := 60  // long enough so that flush comes only from threshold
	maxRetries := 0             // no retries needed
	indexName := "test_index"
	indexer := NewBulkIndexer(threshold, testServer.URL, indexName, flushIntervalSeconds, maxRetries)
	defer indexer.Stop()

	// Create two dummy documents.
	doc1 := &models.Document{
		URL:          "http://example.com/page1",
		CanonicalURL: "",
		Title:        "Page One",
	}
	doc2 := &models.Document{
		URL:          "https://example.com/page2",
		CanonicalURL: "https://example.com/page2",
		Title:        "Page Two",
	}

	// Add documents to the indexer.
	indexer.AddDocumentToIndexerPayload(doc1)
	indexer.AddDocumentToIndexerPayload(doc2)

	// Wait for the flush to occur.
	select {
	case payload := <-payloadCh:
		// The NDJSON payload should consist of 2 documents, each with a meta line and a doc line.
		// We'll split the payload by newline.
		scanner := bufio.NewScanner(bytes.NewReader(payload))
		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) != "" {
				lines = append(lines, line)
			}
		}
		expectedLines := threshold * 2
		if len(lines) != expectedLines {
			t.Errorf("Expected %d NDJSON lines (2 per document), got %d", expectedLines, len(lines))
		}

		// Optionally, decode and verify one meta line.
		var meta map[string]map[string]string
		if err := json.Unmarshal([]byte(lines[0]), &meta); err != nil {
			t.Errorf("Failed to unmarshal meta line: %v", err)
		}
		if meta["index"]["_index"] != indexName {
			t.Errorf("Expected _index to be %q, got %q", indexName, meta["index"]["_index"])
		}
	case <-time.After(3 * time.Second):
		t.Error("Timed out waiting for flush payload")
	}
}

// Verifies that the retry mechanism is exercised when the simulated
// Elasticsearch endpoint returns error codes.
func TestBulkIndexerRetry(t *testing.T) {
	var attemptCount int32 // use atomic counter

	// Create a test server that returns HTTP 500 for the first two attempts,
	// then returns HTTP 200.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer testServer.Close()

	// Use a threshold of 1 so that flush is triggered immediately.
	threshold := 1
	flushIntervalSeconds := 60 // long flush interval; threshold triggers flush
	maxRetries := 3            // allow up to 3 attempts
	indexName := "retry_index"
	indexer := NewBulkIndexer(threshold, testServer.URL, indexName, flushIntervalSeconds, maxRetries)
	defer indexer.Stop()

	// Create a dummy document.
	doc := &models.Document{
		URL:   "http://example.com/retry",
		Title: "Retry Test",
	}

	// Add the document to trigger flush.
	indexer.AddDocumentToIndexerPayload(doc)

	// Wait enough time for the retries to complete.
	time.Sleep(5 * time.Second)

	// Verify that at least 3 attempts were made.
	if atomic.LoadInt32(&attemptCount) < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attemptCount)
	}
}

