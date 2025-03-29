package administrator

import (
    "time"
    "encoding/json"
    "encoding/gob"
    "net/http"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"
    "indexer/internal/pkg/logger"
    "indexer/internal/pkg/models"
)

// Starts the HTTP ingestion service. This is a simple HTTP server that 
// listens for incoming page data and provides a /health endpoint for monitoring.
func startIngestHTTP(admin *administrator, port string) {
    http.HandleFunc("/index", func(writer http.ResponseWriter, request *http.Request) {
        var pageData models.PageData

        contentType := request.Header.Get("Content-Type")
        if contentType != "application/gob" && contentType != "application/octet-stream" {
            http.Error(writer, "expected Content-Type: application/gob", http.StatusUnsupportedMediaType)
            logger.Log.Warn("Unsupported Content-Type", zap.String("content_type", contentType))
            return
        }

        if err := gob.NewDecoder(request.Body).Decode(&pageData); err != nil {
            http.Error(writer, "failed to decode request", http.StatusBadRequest)
            logger.Log.Warn("Failed to decode incoming GOB", zap.Error(err))
            return
        }

        if err := admin.EnqueuePageData(request.Context(), pageData); err != nil {
            http.Error(writer, "failed to enqueue page data", http.StatusInternalServerError)
            logger.Log.Error("Failed to enqueue page data", zap.Error(err))
            return
        }
        writer.WriteHeader(http.StatusAccepted)
        writer.Write([]byte("Page data enqueued"))
    })

    // /metrics endpoint for Prometheus
    http.Handle("/metrics", promhttp.Handler())

    // /health endpoint
    http.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
        health := struct {
            Status     string    `json:"status"`
            QueueDepth int       `json:"queue_depth"`
            Workers    int       `json:"workers"`
            Uptime     string    `json:"uptime"`
            StartTime  time.Time `json:"start_time"`
        }{
            Status:     "OK",
            QueueDepth: admin.QueueDepth(),
            Workers:    admin.WorkerCount(),
            Uptime:     time.Since(admin.StartTime()).String(),
            StartTime:  admin.StartTime(),
        }

        writer.Header().Set("Content-Type", "application/json")
        json.NewEncoder(writer).Encode(health)
    })

    logger.Log.Info("HTTP ingestion service listening", zap.String("address", ":" + port))

    if err := http.ListenAndServe(":" + port, nil); err != nil {
        logger.Log.Fatal("Failed to start ingestion service", zap.Error(err))
    }
}
