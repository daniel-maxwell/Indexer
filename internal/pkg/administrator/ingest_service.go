package administrator

import (
    "encoding/gob"
    "indexer/internal/logger"
    "indexer/internal/pkg/models"
    "net/http"
    "go.uber.org/zap"
)

func startIngestHTTP(admin *administrator, port string) {
    http.HandleFunc("/index", func(writer http.ResponseWriter, request *http.Request) {
        var pageData models.PageData

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

    logger.Log.Info("HTTP ingestion service listening", zap.String("address", ":"+port))

    if err := http.ListenAndServe(":"+port, nil); err != nil {
        logger.Log.Fatal("Failed to start ingestion service", zap.Error(err))
    }
}
