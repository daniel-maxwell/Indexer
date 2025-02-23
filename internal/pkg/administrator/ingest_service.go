// Will listen for incoming requests, respond quickly, then use GOB encoding to deserialize PageData and enqueue it.
package administrator

import (
    "log"
    "net/http"
	"encoding/gob"
	"indexer/internal/pkg/models"
)

// Starts the HTTP server for ingesting page data.
func (admin *administrator) StartService() {
	http.HandleFunc("/index", func(writer http.ResponseWriter, request *http.Request) {
		var pageData models.PageData
		// Decode the GOB-encoded PageData.
		if err := gob.NewDecoder(request.Body).Decode(&pageData); err != nil {
			http.Error(writer, "failed to decode request", http.StatusBadRequest)
			return
		}

		// Enqueue the decoded page data.
		if err := admin.EnqueuePageData(request.Context(), pageData); err != nil {
			http.Error(writer, "failed to enqueue page data", http.StatusInternalServerError)
			return
		}
		writer.WriteHeader(http.StatusAccepted)
		writer.Write([]byte("Page data enqueued"))
	})

	log.Println("Starting HTTP ingestion service on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("failed to start service: %v", err)
	}
}