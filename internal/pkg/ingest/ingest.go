package ingest

import (
    "encoding/json"
    "io"
    "net/http"
    "indexer/internal/pkg/queue"
)

// Returns an HTTP handler function that accepts JSON posts.
func IngestHandler(q *queue.Queue) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
            return
        }
        body, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "Error reading request body", http.StatusBadRequest)
            return
        }
        defer r.Body.Close()

        // Validate JSON structure - TODO: Verify this is necessary?
        var payload map[string]interface{}
        if err := json.Unmarshal(body, &payload); err != nil {
            http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
            return
        }
		
		err = q.Insert(body)
        if err != nil {
            http.Error(w, "Queue is full, try again later", http.StatusServiceUnavailable)
            return
        }

        w.WriteHeader(http.StatusAccepted)
        w.Write([]byte("Accepted"))
    }
}
