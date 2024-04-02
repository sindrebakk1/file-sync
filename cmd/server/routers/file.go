package routers

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"server/controllers"
	"server/services"
)

func NewFileRouter() *http.ServeMux {
	fileService, err := services.NewFileService()
	if err != nil {
		log.Fatal("Failed to create file service:", err)
	}
	fileController := controllers.NewFileController(fileService)

	r := http.NewServeMux()

	r.HandleFunc("POST /{id}", fileController.GetStatusHandler)
	r.HandleFunc("GET /{id}/download", fileController.DownloadFileHandler)
	r.HandleFunc("GET /{id}/upload", fileController.GetUploadSessionHandler)
	r.HandleFunc("POST /{id}/upload/{sessionId}", fileController.UploadChunkHandler)
	r.HandleFunc("POST /{id}/upload/{sessionId}/commit", fileController.CommitChunksHandler)
	r.HandleFunc("GET /", fileController.GetSyncedFileMapHandler)

	return r
}
