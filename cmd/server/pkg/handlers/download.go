package handlers

import (
	"file-sync/models"
	log "github.com/sirupsen/logrus"
	"server/pkg/mux"
)

// HandleDownload is a mux.HandlerFunc
func HandleDownload(resChan chan models.Message, req *mux.Request) error {
	log.Debug("HandleDownload")
	return nil
}
