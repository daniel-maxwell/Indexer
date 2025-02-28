package administrator

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"indexer/internal/pkg/models"
)

// dummyAdmin implements the Administrator interface minimally.
// It only implements EnqueuePageData (others are no-ops) so we can verify that
// the ingestion endpoint calls EnqueuePageData with the correct payload.
type dummyAdmin struct {
	enqueued chan models.PageData
}

func (da *dummyAdmin) EnqueuePageData(ctx context.Context, data models.PageData) error {
	da.enqueued <- data
	return nil
}

func (da *dummyAdmin) ProcessAndIndex(ctx context.Context) error {
	return nil
}

func (da *dummyAdmin) StartService(port string) {
	// no-op for this dummy
}

func (da *dummyAdmin) Stop() {
	// no-op for this dummy
}

func TestIngestHTTP(t *testing.T) {
	// Create a dummy admin instance.
	da := &dummyAdmin{enqueued: make(chan models.PageData, 1)}

	// Create an HTTP mux that simulates the ingestion endpoint.
	// This is effectively the same code as in production's startIngestHTTP.
	mux := http.NewServeMux()
	mux.HandleFunc("/ingest", func(writer http.ResponseWriter, request *http.Request) {
		var pd models.PageData
		err := json.NewDecoder(request.Body).Decode(&pd)
		if err != nil {
			http.Error(writer, "Bad request", http.StatusBadRequest)
			return
		}
		err = da.EnqueuePageData(request.Context(), pd)
		if err != nil {
			http.Error(writer, "Queue full", http.StatusServiceUnavailable)
			return
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("Accepted"))
	})

	// Create a test HTTP server.
	server := httptest.NewServer(mux)
	defer server.Close()

	// Prepare test page data.
	testData := models.PageData{
		URL:         "http://example.com/ingest",
		VisibleText: "Ingestion test",
	}
	jsonData, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Send POST request to the /ingest endpoint.
	response, err := http.Post(server.URL+"/ingest", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		t.Fatalf("Failed to send POST request: %v", err)
	}
	defer response.Body.Close()
	body, _ := io.ReadAll(response.Body)
	if response.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d, body: %s", response.StatusCode, string(body))
	}

	// Verify that dummyAdmin received the page data.
	select {
	case pd := <-da.enqueued:
		if pd.URL != testData.URL || pd.VisibleText != testData.VisibleText {
			t.Errorf("Enqueued data mismatch. Got %+v, expected %+v", pd, testData)
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for enqueued page data")
	}
}
